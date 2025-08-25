# GEMINI.md - Natural Hazard Disclosure (NHD) Service

This file provides a comprehensive overview of the Natural Hazard Disclosure (NHD) service project, designed to be used as a context for AI-powered development tools like Gemini.

## Project Overview

This project is a cloud-native application designed to generate Natural Hazard Disclosure (NHD) reports for real estate transactions in California. The system is built on Google Cloud Platform (GCP) and utilizes a serverless, event-driven architecture.

**Key Technologies:**

*   **Frontend:** Single-page application (SPA) using a modern JavaScript framework (e.g., React or Vue), hosted on Firebase Hosting.
*   **Backend API:** A stateless microservice written in Go and deployed on Cloud Run.
*   **Report Generation:** A Python Cloud Function triggered by Cloud Pub/Sub messages.
*   **Database:** Google Cloud Firestore (NoSQL).
*   **Storage:** Google Cloud Storage for PDF reports.
*   **Authentication:** Firebase Authentication.
*   **Email:** SendGrid.
*   **Data Format:** Protocol Buffers (Protobufs) are used for data consistency between services.

**Architecture:**

The system follows a microservices architecture:

1.  A user interacts with the frontend to request an NHD report.
2.  The Go backend API receives the request, creates a `ReportRun` document in Firestore with a "PENDING" status, and publishes a message to a Pub/Sub topic.
3.  The Python Report Generation service is triggered by the Pub/Sub message, performs the geospatial analysis, generates a PDF report, stores it in Cloud Storage, and updates the Firestore document to "COMPLETED".

## Building and Running

### Go Backend

To run the backend server:
```bash
cd backend
export GOOGLE_CLOUD_PROJECT=<your-gcp-project-id>
go run main.go
```

### Python Report Generation Service

To deploy the Cloud Function:
```bash
gcloud functions deploy handle_report_request \
    --runtime python312 \
    --trigger-topic nhd-report-requests \
    --source reporter
```

### Frontend

*TODO: Add commands for building and running the frontend*
```bash
# npm install
# npm start
```

## Development Conventions

*   **Protocol Buffers:** Protobufs are the canonical data format. When adding or modifying data structures, update the `.proto` definitions first.
*   **Testing:** The project uses a multi-layered testing strategy:
    *   **Unit Tests (Go):** Focused on individual functions and methods using `testify/assert` and `testify/mock`.
    *   **Integration Tests (Go):** Using `Testcontainers` to test interactions with Firestore and Pub/Sub emulators.
    *   **End-to-End (E2E) Tests:** Using a browser automation framework like Cypress or Playwright to validate the entire workflow.
*   **Logging:** Both Go and Python services should output structured JSON logs to Cloud Logging for better observability.
*   **API:** The API is documented in the `README.md` file. Follow the existing RESTful conventions when adding new endpoints.
