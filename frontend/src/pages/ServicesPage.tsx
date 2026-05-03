import { useEffect, useMemo, useState, type ChangeEvent } from "react";

import { getCartIcon, getServices } from "../api/publicApi";
import { ServiceCard } from "../components/services/ServiceCard";
import { ServicesFilters, type ServiceFilterState } from "../components/services/ServicesFilters";
import type { Claim, Service } from "../types/domain";
import { pluralizeServices } from "../utils/format";

interface ServicesPageProps {
  services: Service[];
  draftClaim: Claim;
  onAddService: (serviceSlug: string) => void;
}

const initialFilters: ServiceFilterState = {
  query: ""
};

export const ServicesPage = ({ services, draftClaim, onAddService }: ServicesPageProps): JSX.Element => {
  const [filters, setFilters] = useState<ServiceFilterState>(initialFilters);
  const [appliedFilters, setAppliedFilters] = useState<ServiceFilterState>(initialFilters);
  const [items, setItems] = useState<Service[]>(services);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [servicesNote, setServicesNote] = useState<string>("");
  const [cartCount, setCartCount] = useState<number>(0);

  const [imageFile, setImageFile] = useState<File | null>(null);
  const [clipThreshold, setClipThreshold] = useState<number>(0.55);
  const [clipTopK, setClipTopK] = useState<number>(5);
  const [clipSearchActive, setClipSearchActive] = useState<boolean>(false);
  const [clipSearchLoading, setClipSearchLoading] = useState<boolean>(false);
  const [clipError, setClipError] = useState<string>("");
  const [clipScores, setClipScores] = useState<Map<number, number>>(new Map());
  const [clipServices, setClipServices] = useState<Service[]>([]);

  useEffect(() => {
    let isMounted = true;

    const loadServices = async (): Promise<void> => {
      setIsLoading(true);
      const response = await getServices(appliedFilters);
      if (!isMounted) {
        return;
      }

      setItems(response.data);
      setServicesNote(response.note ?? "");
      setIsLoading(false);
      setClipSearchActive(false);
      setClipScores(new Map());
      setClipServices([]);
      setClipError("");
    };

    void loadServices();

    return () => {
      isMounted = false;
    };
  }, [appliedFilters]);

  useEffect(() => {
    let isMounted = true;

    const loadCartIcon = async (): Promise<void> => {
      const response = await getCartIcon();
      if (!isMounted) {
        return;
      }

      const localDraftCount = draftClaim.rows.reduce((sum, row) => sum + row.quantity, 0);
      setCartCount(response.source === "backend" ? response.data.serviceCount : localDraftCount);
    };

    void loadCartIcon();

    return () => {
      isMounted = false;
    };
  }, [draftClaim.rows]);

  const shownServices = useMemo(() => {
    if (!clipSearchActive) {
      return items;
    }
    return clipServices;
  }, [clipSearchActive, clipServices, items]);

  const handleImageChange = (event: ChangeEvent<HTMLInputElement>): void => {
    const file = event.target.files?.[0] ?? null;
    setImageFile(file);
    setClipError("");
  };

  const handleTopKChange = (event: ChangeEvent<HTMLInputElement>): void => {
    const parsed = Number(event.target.value);
    if (Number.isFinite(parsed)) {
      setClipTopK(parsed);
    } else {
      setClipTopK(5);
    }
  };

  const handleClipSearch = async (): Promise<void> => {
    if (!imageFile) {
      setClipError("Выберите изображение для CLIP-поиска.");
      return;
    }

    setClipSearchLoading(true);
    setClipError("");

    try {
      const { searchByImageWithClipDetailed } = await import("../modules/clipSearch");
      const detailed = await searchByImageWithClipDetailed(items, imageFile, clipThreshold, clipTopK);

      const scoreMap = new Map<number, number>();
      const matchedServices = detailed.matches.map((result) => {
        scoreMap.set(result.service.id, result.score);
        return result.service;
      });

      setClipScores(scoreMap);
      setClipServices(matchedServices);
      setClipSearchActive(true);

      if (matchedServices.length === 0) {
        const maxScore = detailed.allScores.length > 0 ? detailed.allScores[0].score.toFixed(3) : "0.000";
        setClipError(
          `Совпадений не найдено при threshold=${detailed.threshold.toFixed(2)}. Максимальный score=${maxScore}. Снизьте threshold.`
        );
      }
    } catch {
      setClipError("Не удалось выполнить CLIP-поиск. Проверьте подключение и повторите попытку.");
    } finally {
      setClipSearchLoading(false);
    }
  };

  const resetClipSearch = (): void => {
    setClipSearchActive(false);
    setClipScores(new Map());
    setClipServices([]);
    setClipError("");
  };

  return (
    <>
      <section className="list-actions">
        <ServicesFilters value={filters} onChange={setFilters} onSubmit={() => setAppliedFilters(filters)} />

        <a
          className={`cart-link ${cartCount === 0 ? "cart-link-disabled" : ""}`}
          href="/api/claims/cart-icon"
          target="_blank"
          rel="noreferrer"
          title="Открыть ответ корзины (JSON)"
        >
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="M12 3c-1.2 0-2.2.8-2.5 2H8.4c-.8 0-1.4.6-1.4 1.4V7h10V6.4c0-.8-.6-1.4-1.4-1.4h-1.1C14.2 3.8 13.2 3 12 3zm-2.8 4v2.2c0 1.9-1.2 3.6-3.1 4.3l1 5c.1.9.8 1.5 1.7 1.5h6.4c.9 0 1.6-.6 1.7-1.5l1-5c-1.9-.7-3.1-2.4-3.1-4.3V7h-5.6zm2.8 2.3c.5 0 .9.4.9.9v3.6c0 .5-.4.9-.9.9s-.9-.4-.9-.9v-3.6c0-.5.4-.9.9-.9z" />
          </svg>
          <span>Иконка корзины</span>
          <strong>
            {cartCount} {pluralizeServices(cartCount)}
          </strong>
        </a>
      </section>

      <section className="card clip-panel">
        <div className="clip-grid">
          <input className="search-input" type="file" accept="image/*" onChange={handleImageChange} />
          <label className="clip-control">
            Threshold ({clipThreshold.toFixed(2)})
            <input
              type="range"
              min={0.4}
              max={0.9}
              step={0.01}
              value={clipThreshold}
              onChange={(event) => setClipThreshold(Number(event.target.value))}
            />
          </label>
          <label className="clip-control">
            TopK
            <input
              className="search-input"
              type="number"
              min={1}
              max={20}
              value={clipTopK}
              onChange={handleTopKChange}
            />
          </label>
          <button className="chip chip-button" type="button" onClick={handleClipSearch} disabled={clipSearchLoading}>
            {clipSearchLoading ? "Поиск..." : "Найти по изображению"}
          </button>
          <button className="chip chip-button" type="button" onClick={resetClipSearch}>
            Сброс CLIP-фильтра
          </button>
        </div>
        {clipError.length > 0 && <p className="notice warn">{clipError}</p>}
      </section>

      {servicesNote.length > 0 && <p className="notice warn">{servicesNote}</p>}

      <section className="services-grid">
        {isLoading ? (
          <article className="card">Загрузка списка услуг...</article>
        ) : shownServices.length > 0 ? (
          shownServices.map((service) => (
            <ServiceCard
              key={service.id}
              service={service}
              onAddService={onAddService}
              clipScore={clipScores.get(service.id)}
            />
          ))
        ) : (
          <article className="card">По вашему запросу ничего не найдено.</article>
        )}
      </section>
    </>
  );
};
