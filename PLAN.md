# Plan: Commercial-Grade NHD Service Evolution

This plan outlines the steps to transform the current NHD service prototype into a production-ready, commercial-grade platform.

## Architectural Report: Current State vs. Commercial Standards

### Current State Analysis
- **Orchestration**: Go backend on Cloud Run is solid. It uses Firestore for state and Pub/Sub for async processing.
- **Geospatial Engine**: Python Cloud Function uses GeoJSON files. While functional for a few points, this will not scale to production volumes or handle complex multi-layer analysis efficiently.
- **Data Depth**: Only covers the 6 statutory hazards. Commercial reports require 20+ layers (Local, Tax, Environmental).
- **Output**: No physical PDF generation or storage implemented.
- **Financials**: Payment status is manually updated; no real gateway integration.
- **Address Handling**: Relies on user-provided coordinates; no rigorous validation or geocoding.

### Recommended Changes
1. **Geospatial Infrastructure**: Move from GeoJSON files to a managed spatial database (Cloud SQL with PostGIS). This enables efficient querying of large datasets and simplifies adding new layers.
2. **Data Model Expansion**: Update `nhd.proto` to include a full suite of CA disclosures (Supplemental Hazards, Property Taxes, Environmental Hazards).
3. **PDF Generation Engine**: Implement a robust PDF pipeline in the Python service using a library like `ReportLab`. Must include the legally mandated Statutory Summary Page.
4. **Integration Services**:
    - **Google Places API**: For address standardization and high-precision geocoding.
    - **Stripe**: For automated payment processing and webhooks.
    - **SendGrid**: Finalize the implementation for automated report delivery.
5. **Observability**: Implement trace IDs across Go/Python boundaries to debug specific report runs effectively.

---

## Milestone Plan

All milestones follow the requirement: **Compile/Test/Lint + E2E + Manual Verification.**

### Milestone 1: Geospatial Core & Expanded Data Model
- **Goal**: Establish the production data foundation.
- **Tasks**:
    1. Update `nhd.proto` to include `SupplementalResults`, `TaxResults`, and `EnvironmentalResults`.
    2. Provision and configure Cloud SQL (PostGIS).
    3. Implement a data ingestion pipeline for real CA shapefiles (FEMA, CAL FIRE, CGS).
    4. Refactor `reporter/main.py` to perform PIP analysis against PostGIS.
- **Verification**: 
    - Unit tests for new proto definitions.
    - Integration tests verifying PIP analysis against PostGIS for known coordinates.
    - `make test` & `make lint`.

### Milestone 2: Statutory PDF Generation & GCS Integration
- **Goal**: Produce the actual legal document.
- **Tasks**:
    1. Implement PDF template engine in `reporter/` (ReportLab + Jinja2).
    2. Generate a multi-page PDF including the Cover, Statutory Summary, and detailed sections.
    3. Upload generated PDFs to Google Cloud Storage.
    4. Update Go API to generate Signed URLs for secure frontend downloads.
- **Verification**: 
    - Manual inspection of generated PDF formatting.
    - Integration tests for GCS upload/download.
    - E2E test verifying a "COMPLETED" report has a valid download link.

### Milestone 3: Professional Address Validation & Geocoding
- **Goal**: Ensure analysis is performed on the correct location.
- **Tasks**:
    1. Integrate Google Places API in `backend/api`.
    2. Update `POST /report-runs` to validate the address and fetch precise coordinates before creating the run.
    3. Update Frontend with an address autocomplete UI.
- **Verification**:
    - Manual verification of geocoding accuracy for edge-case addresses.
    - E2E test for the report submission flow.

### Milestone 4: Automated Financials & Stripe Integration
- **Goal**: Enable self-service report purchasing.
- **Tasks**:
    1. Integrate Stripe Go SDK.
    2. Implement `POST /report-runs/{id}/create-checkout-session` endpoint.
    3. Implement Stripe Webhook handler to transition report status to `PROCESSING` only after payment success.
    4. Update Frontend to handle Stripe redirection and success/cancel flows.
- **Verification**:
    - Successful test payments in Stripe sandbox.
    - E2E test: User pays -> Pub/Sub message sent -> Report generated.

### Milestone 5: Final Hardening, Monitoring & Delivery
- **Goal**: Production launch readiness.
- **Tasks**:
    1. Finalize SendGrid integration in the Python service.
    2. Implement trace-id logging across services.
    3. Set up Cloud Monitoring dashboards for API latency and report generation success rates.
    4. Perform comprehensive `make test-e2e` against a suite of "golden" properties.
- **Verification**:
    - All tests pass (`make clean`, `make test`, `make test-e2e`, `make lint`).
    - Manual verification of the full "Order to Email" loop.
