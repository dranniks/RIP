import { env, pipeline } from "@huggingface/transformers";

import type { Service } from "../types/domain";

const MODEL_ID = "Xenova/siglip-base-patch16-224";

type ZeroShotScore = {
  label: string;
  score: number;
};

type ZeroShotClassifier = (
  image: Blob,
  candidateLabels: string[],
  options?: { hypothesis_template?: string }
) => Promise<ZeroShotScore[] | ZeroShotScore[][]>;

interface ClipRuntime {
  classifier: ZeroShotClassifier;
}

type MetalKind = "bronze" | "brass" | "iron" | "silver" | "unknown";

interface VisualProfile {
  brightness: number;
  saturation: number;
  warmth: number;
  highlights: number;
}

export interface ClipSearchResult {
  service: Service;
  score: number;
}

export interface ClipServiceScore {
  service: Service;
  description: string;
  rawModelScore: number;
  rawAggregateScore: number;
  visualScore: number;
  clipWeight: number;
  visualWeight: number;
  score: number;
}

export interface ClipSearchDetailedResult {
  matches: ClipSearchResult[];
  allScores: ClipServiceScore[];
  threshold: number;
  topK: number;
}

let runtimePromise: Promise<ClipRuntime> | null = null;

const clamp = (value: number, min: number, max: number): number => Math.min(max, Math.max(min, value));

const angularDistance = (left: number, right: number): number => {
  const diff = Math.abs(left - right) % 360;
  return Math.min(diff, 360 - diff);
};

const warmnessFromHue = (hueDeg: number, saturation: number): number => {
  if (saturation < 0.12) {
    return 0.2;
  }

  const distToBrown = angularDistance(hueDeg, 28);
  const distToYellow = angularDistance(hueDeg, 55);
  const distToOrange = angularDistance(hueDeg, 38);
  const best = Math.min(distToBrown, distToYellow, distToOrange);

  return clamp(1 - best / 180, 0, 1);
};

const rgbToHsv = (r: number, g: number, b: number): { h: number; s: number; v: number } => {
  const max = Math.max(r, g, b);
  const min = Math.min(r, g, b);
  const delta = max - min;

  let hue = 0;
  if (delta > 1e-9) {
    if (max === r) {
      hue = ((g - b) / delta) % 6;
    } else if (max === g) {
      hue = (b - r) / delta + 2;
    } else {
      hue = (r - g) / delta + 4;
    }

    hue *= 60;
    if (hue < 0) {
      hue += 360;
    }
  }

  const saturation = max <= 1e-9 ? 0 : delta / max;
  return {
    h: hue,
    s: clamp(saturation, 0, 1),
    v: clamp(max, 0, 1)
  };
};

const sampleVisualProfile = (pixels: Uint8ClampedArray, skipBlackBackground: boolean): VisualProfile | null => {
  let brightnessSum = 0;
  let saturationSum = 0;
  let warmthSum = 0;
  let highlights = 0;
  let samples = 0;

  for (let index = 0; index < pixels.length; index += 4) {
    const alpha = pixels[index + 3] / 255;
    if (alpha < 0.2) {
      continue;
    }

    const r = pixels[index] / 255;
    const g = pixels[index + 1] / 255;
    const b = pixels[index + 2] / 255;

    if (skipBlackBackground && r < 0.08 && g < 0.08 && b < 0.08) {
      continue;
    }

    const hsv = rgbToHsv(r, g, b);

    brightnessSum += hsv.v;
    saturationSum += hsv.s;
    warmthSum += warmnessFromHue(hsv.h, hsv.s);

    if (hsv.v >= 0.7 && hsv.s <= 0.28) {
      highlights += 1;
    }
    samples += 1;
  }

  if (samples === 0) {
    return null;
  }

  return {
    brightness: brightnessSum / samples,
    saturation: saturationSum / samples,
    warmth: warmthSum / samples,
    highlights: highlights / samples
  };
};

const extractVisualProfile = async (image: Blob): Promise<VisualProfile | null> => {
  if (typeof document === "undefined" || typeof createImageBitmap === "undefined") {
    return null;
  }

  try {
    const bitmap = await createImageBitmap(image);
    const targetWidth = Math.max(1, Math.min(224, bitmap.width));
    const targetHeight = Math.max(1, Math.round((bitmap.height / bitmap.width) * targetWidth));

    const canvas = document.createElement("canvas");
    canvas.width = targetWidth;
    canvas.height = targetHeight;

    const context = canvas.getContext("2d", { willReadFrequently: true });
    if (!context) {
      bitmap.close();
      return null;
    }

    context.drawImage(bitmap, 0, 0, targetWidth, targetHeight);
    bitmap.close();

    const pixels = context.getImageData(0, 0, targetWidth, targetHeight).data;
    const focusedProfile = sampleVisualProfile(pixels, true);

    if (focusedProfile) {
      return focusedProfile;
    }

    return sampleVisualProfile(pixels, false);
  } catch {
    return null;
  }
};

const promptVariantsBySlug: Record<string, string[]> = {
  "alloy-bronze-cyprus": [
    "dark brown bronze ingot",
    "oxidized bronze metal bar",
    "copper tin bronze artifact",
    "ancient bronze alloy ingot"
  ],
  "bronze-cyprus": [
    "dark brown bronze ingot",
    "oxidized bronze metal bar",
    "copper tin bronze artifact",
    "ancient bronze alloy ingot"
  ],
  "alloy-brass-rome": [
    "warm yellow brass ingot",
    "golden brass metal bar",
    "copper zinc brass alloy",
    "roman brass artifact ingot"
  ],
  "brass-rome": [
    "warm yellow brass ingot",
    "golden brass metal bar",
    "copper zinc brass alloy",
    "roman brass artifact ingot"
  ],
  "alloy-iron-north": [
    "dark gray iron ingot",
    "matte iron metal bar",
    "rough black iron alloy",
    "forged iron artifact bar"
  ],
  "iron-north": [
    "dark gray iron ingot",
    "matte iron metal bar",
    "rough black iron alloy",
    "forged iron artifact bar"
  ],
  "alloy-silver-byzantium": [
    "bright silver ingot",
    "pale silver metal bar",
    "shiny silver alloy artifact",
    "high reflectance silver bar"
  ],
  "silver-byzantium": [
    "bright silver ingot",
    "pale silver metal bar",
    "shiny silver alloy artifact",
    "high reflectance silver bar"
  ]
};

const ensureRuntime = async (): Promise<ClipRuntime> => {
  if (!runtimePromise) {
    env.allowLocalModels = false;
    env.useBrowserCache = true;

    runtimePromise = pipeline("zero-shot-image-classification", MODEL_ID).then((classifier) => ({
      classifier: classifier as ZeroShotClassifier
    }));
  }

  return runtimePromise;
};

const unwrapScores = (output: ZeroShotScore[] | ZeroShotScore[][]): ZeroShotScore[] => {
  if (!Array.isArray(output) || output.length === 0) {
    return [];
  }

  const first = output[0];
  if (Array.isArray(first)) {
    return first as ZeroShotScore[];
  }

  return output as ZeroShotScore[];
};

const normalizeSlug = (slug: string): string => slug.trim().toLowerCase();

const inferMetalKind = (service: Service): MetalKind => {
  const haystack = `${service.slug} ${service.name} ${service.clipDescriptionEn}`.toLowerCase();

  if (haystack.includes("bronze") || haystack.includes("бронз")) {
    return "bronze";
  }
  if (haystack.includes("brass") || haystack.includes("латун")) {
    return "brass";
  }
  if (haystack.includes("iron") || haystack.includes("желез")) {
    return "iron";
  }
  if (haystack.includes("silver") || haystack.includes("сереб")) {
    return "silver";
  }

  return "unknown";
};

const serviceVisualProfiles: Record<MetalKind, VisualProfile> = {
  bronze: { brightness: 0.38, saturation: 0.44, warmth: 0.78, highlights: 0.15 },
  brass: { brightness: 0.62, saturation: 0.42, warmth: 0.9, highlights: 0.28 },
  iron: { brightness: 0.34, saturation: 0.09, warmth: 0.26, highlights: 0.08 },
  silver: { brightness: 0.74, saturation: 0.08, warmth: 0.32, highlights: 0.42 },
  unknown: { brightness: 0.52, saturation: 0.25, warmth: 0.48, highlights: 0.2 }
};

const scoreVisualSimilarity = (queryProfile: VisualProfile | null, metalKind: MetalKind): number => {
  if (!queryProfile) {
    return 0.5;
  }

  const target = serviceVisualProfiles[metalKind];

  let distance =
    Math.abs(queryProfile.brightness - target.brightness) * 0.48 +
    Math.abs(queryProfile.saturation - target.saturation) * 0.22 +
    Math.abs(queryProfile.warmth - target.warmth) * 0.2 +
    Math.abs(queryProfile.highlights - target.highlights) * 0.1;

  if (queryProfile.saturation < 0.14 && queryProfile.brightness < 0.52 && metalKind === "silver") {
    distance += 0.12;
  }
  if (queryProfile.saturation < 0.14 && queryProfile.brightness < 0.52 && metalKind === "iron") {
    distance -= 0.06;
  }

  return clamp(1 - distance, 0, 1);
};

const inferPromptFamily = (slug: string): string[] => {
  const lower = normalizeSlug(slug);

  if (lower.includes("silver")) {
    return ["bright silver ingot", "shiny silver metal bar", "pale reflective silver artifact"];
  }

  if (lower.includes("brass")) {
    return ["warm yellow brass ingot", "golden brass metal bar", "copper zinc brass artifact"];
  }

  if (lower.includes("bronze")) {
    return ["dark brown bronze ingot", "oxidized bronze metal bar", "copper tin bronze artifact"];
  }

  if (lower.includes("iron")) {
    return ["dark gray iron ingot", "matte iron metal bar", "forged black iron artifact"];
  }

  return ["archaeological metal alloy ingot", "ancient metal artifact bar"];
};

const buildServicePrompts = (service: Service): string[] => {
  const bySlug = promptVariantsBySlug[normalizeSlug(service.slug)] ?? inferPromptFamily(service.slug);
  const fromDescription = service.clipDescriptionEn.trim();

  const prompts = [...bySlug];
  if (fromDescription.length > 0) {
    prompts.push(fromDescription);
  }

  // Keep uniqueness and filter obviously broken strings.
  return [...new Set(prompts.map((item) => item.trim()).filter((item) => item.length >= 8))];
};

const averageTop = (values: number[], count: number): number => {
  if (values.length === 0) {
    return 0;
  }
  const sorted = [...values].sort((a, b) => b - a);
  const top = sorted.slice(0, Math.max(1, Math.min(count, sorted.length)));
  const sum = top.reduce((acc, item) => acc + item, 0);
  return sum / top.length;
};

export const searchByImageWithClipDetailed = async (
  services: Service[],
  imageFile: File,
  threshold: number,
  topK: number
): Promise<ClipSearchDetailedResult> => {
  const candidates = services.filter((service) => service.clipDescriptionEn.trim().length > 0);
  if (candidates.length === 0) {
    return {
      matches: [],
      allScores: [],
      threshold: 0.55,
      topK: 5
    };
  }

  const runtime = await ensureRuntime();

  const servicePrompts = new Map<number, string[]>();
  const allPrompts: string[] = [];
  for (const service of candidates) {
    const prompts = buildServicePrompts(service);
    servicePrompts.set(service.id, prompts);
    allPrompts.push(...prompts);
  }

  // Deduplicate globally for one model call.
  const uniquePrompts = [...new Set(allPrompts)];

  const output = await runtime.classifier(imageFile, uniquePrompts, {
    hypothesis_template: "This is a photo of {}"
  });
  const ranked = unwrapScores(output);
  const promptScoreMap = new Map<string, number>(ranked.map((item) => [item.label, item.score]));

  const safeThreshold = Number.isFinite(threshold) ? threshold : 0.55;
  const safeTopK = Number.isFinite(topK) ? topK : 5;
  const normalizedThreshold = clamp(safeThreshold, 0.4, 0.9);
  const normalizedTopK = Math.max(1, Math.min(20, Math.trunc(safeTopK)));
  const queryVisualProfile = await extractVisualProfile(imageFile);
  const visualWeight = queryVisualProfile && queryVisualProfile.saturation < 0.16 ? 0.7 : 0.45;
  const clipWeight = 1 - visualWeight;

  const rawScores = candidates.map((service) => {
    const prompts = servicePrompts.get(service.id) ?? [];
    const scoredPrompts = prompts
      .map((prompt) => ({
        prompt,
        value: promptScoreMap.get(prompt) ?? 0
      }))
      .sort((left, right) => right.value - left.value);

    const best = scoredPrompts[0] ?? { prompt: service.clipDescriptionEn, value: 0 };
    const aggregate = averageTop(
      scoredPrompts.map((item) => item.value),
      2
    );
    const visualScore = scoreVisualSimilarity(queryVisualProfile, inferMetalKind(service));

    return {
      service,
      description: best.prompt,
      rawModelScore: best.value,
      rawAggregateScore: aggregate,
      visualScore
    };
  });

  const allScores: ClipServiceScore[] = rawScores
    .map((item) => ({
      ...item,
      clipWeight,
      visualWeight,
      score: clamp(item.rawAggregateScore * clipWeight + item.visualScore * visualWeight, 0, 1)
    }))
    .sort((left, right) => right.score - left.score);

  const matches = allScores
    .filter((item) => item.score >= normalizedThreshold)
    .map((item) => ({
      service: item.service,
      score: item.score
    }))
    .slice(0, normalizedTopK);

  return {
    matches,
    allScores,
    threshold: normalizedThreshold,
    topK: normalizedTopK
  };
};

export const searchByImageWithClip = async (
  services: Service[],
  imageFile: File,
  threshold: number,
  topK: number
): Promise<ClipSearchResult[]> => {
  const detailed = await searchByImageWithClipDetailed(services, imageFile, threshold, topK);
  return detailed.matches;
};
