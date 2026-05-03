import type { Service } from "../types/domain";

export const mockServices: Service[] = [
  {
    id: 1,
    slug: "bronze-cyprus",
    name: "Bronze Cyprus",
    era: "Late Bronze Age",
    culture: "Eastern Mediterranean",
    description: "Эталонный бронзовый сплав Кипра для XRF-сопоставления.",
    clipDescriptionEn: "Dark brown bronze ingot with oxidized rough texture and warm copper highlights under light.",
    imageUrl: "",
    cuReference: "0.83",
    znReference: "0.04",
    snReference: "0.31",
    pbReference: "0.12",
    price: 1250,
    availableDate: "2026-03-05"
  },
  {
    id: 2,
    slug: "brass-rome",
    name: "Brass Rome",
    era: "I-III centuries",
    culture: "Roman Empire",
    description: "Эталонная латунь римского периода для XRF-анализа.",
    clipDescriptionEn: "Warm yellow brass ingot with bright golden edges and medium metallic reflectance on surface.",
    imageUrl: "",
    cuReference: "0.78",
    znReference: "0.64",
    snReference: "0.06",
    pbReference: "0.03",
    price: 1420,
    availableDate: "2026-03-11"
  },
  {
    id: 3,
    slug: "iron-north",
    name: "Железный сплав Севера",
    era: "VIII-XI вв.",
    culture: "Скандинавские мастерские",
    description: "Эталонный железный сплав северной традиции для XRF-сопоставления.",
    clipDescriptionEn: "Dark gray iron ingot with matte grainy texture, cold tone, and weak silver highlights.",
    imageUrl: "",
    cuReference: "0.14",
    znReference: "0.02",
    snReference: "0.01",
    pbReference: "0.01",
    price: 900,
    availableDate: "2026-03-18"
  },
  {
    id: 4,
    slug: "silver-byzantium",
    name: "Silver Byzantium",
    era: "X-XI centuries",
    culture: "Byzantine workshops",
    description: "Эталонный серебряный сплав византийского периода.",
    clipDescriptionEn: "Pale silver ingot with smooth texture, strong reflectance, and bright white specular highlights.",
    imageUrl: "",
    cuReference: "0.29",
    znReference: "0.01",
    snReference: "0",
    pbReference: "0.09",
    price: 2100,
    availableDate: "2026-03-27"
  }
];

export const servicesBySlug = new Map(mockServices.map((service) => [service.slug, service]));
export const servicesByID = new Map(mockServices.map((service) => [service.id, service]));
