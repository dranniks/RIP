import { useState } from "react";
import { Navigate, Route, Routes } from "react-router-dom";

import { AppBreadcrumbs } from "./components/layout/AppBreadcrumbs";
import { SiteHeader } from "./components/layout/SiteHeader";
import { createInitialDraftClaim, createPendingRow } from "./mock/claims";
import { mockServices } from "./mock/services";
import { NotFoundPage } from "./pages/NotFoundPage";
import { ServiceDetailPage } from "./pages/ServiceDetailPage";
import { ServicesPage } from "./pages/ServicesPage";

function App(): JSX.Element {
  const [draftClaim, setDraftClaim] = useState(createInitialDraftClaim);

  const handleAddService = (serviceSlug: string): void => {
    setDraftClaim((current) => {
      const rowIndex = current.rows.findIndex((row) => row.serviceSlug === serviceSlug);
      const rows = [...current.rows];

      if (rowIndex >= 0) {
        rows[rowIndex] = {
          ...rows[rowIndex],
          quantity: rows[rowIndex].quantity + 1
        };
      } else {
        rows.push(createPendingRow(serviceSlug));
      }

      return {
        ...current,
        status: current.status === "удален" ? "черновик" : current.status,
        formedAt: current.status === "удален" ? "-" : current.formedAt,
        completionFormulaResult: current.status === "удален" ? "" : current.completionFormulaResult,
        rows
      };
    });
  };

  return (
    <>
      <SiteHeader />

      <main className="page">
        <AppBreadcrumbs services={mockServices} />

        <Routes>
          <Route path="/" element={<Navigate to="/services" replace />} />
          <Route
            path="/services"
            element={<ServicesPage services={mockServices} draftClaim={draftClaim} onAddService={handleAddService} />}
          />
          <Route path="/services/:id" element={<ServiceDetailPage services={mockServices} />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </main>
    </>
  );
}

export default App;
