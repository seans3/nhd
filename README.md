# **Natural Hazard Disclosure (NHD) Service Specification**

This document outlines the complete architecture, data models, and workflows for a cloud-native application designed to generate Natural Hazard Disclosure (NHD) reports for real estate transactions in California.

## **Overview**

A **Natural Hazard Disclosure (NHD) report** is a legally mandated document required in nearly all California real estate transactions. Its primary purpose is to inform potential buyers about whether a property lies within specific, officially mapped natural hazard zones. This disclosure is a critical component of consumer protection, ensuring that a buyer is aware of significant risks that could affect the property's value, safety, and the cost of insurance before they complete the purchase. 🏡

Under California law (the Natural Hazards Disclosure Act), sellers and their agents have a legal duty to disclose this information. Failure to do so can result in significant legal liability. By providing a standardized report from a third-party expert (like the service described in this document), sellers fulfill their legal obligations and buyers can make a more informed decision.

### **Legally Required Hazard Disclosures**

The report must, by law, disclose whether the property is located within any of the following six statutorily defined hazard zones:

1. **Special Flood Hazard Area**: A federally designated zone with a 1% or greater annual chance of flooding. A property in this zone typically requires mandatory flood insurance.  
2. **Dam Inundation Area**: An area that would be flooded in the event of a dam failure.  
3. **Very High Fire Hazard Severity Zone**: A zone with a very high risk of wildfire damage, as mapped by CAL FIRE. Properties here may face higher insurance premiums and stricter building codes. 🔥  
4. **Wildland Fire Area**: An area where the state has the primary financial responsibility for preventing and suppressing fires, indicating significant wildfire risk.  
5. **Earthquake Fault Zone**: An area directly over or very near an active earthquake fault. Construction on or near these faults is heavily restricted.  
6. **Seismic Hazard Zone**: An area at risk of secondary earthquake effects like **liquefaction** (where soil loses strength and behaves like a liquid) or **earthquake-induced landslides**.

## **High-Level GCP Architecture**

The application is designed as a modern, scalable, and secure cloud-native system built entirely on Google Cloud Platform (GCP). It uses a serverless, event-driven architecture to separate the user-facing API from the resource-intensive data processing, ensuring a responsive user experience and minimizing operational costs.

The architecture consists of several key components:

* **Frontend**: A single-page application (SPA) built with a modern JavaScript framework (like React or Vue) and hosted on **Firebase Hosting** for global CDN delivery.  
* **Backend API (Go)**: A stateless microservice written in Go and deployed on **Cloud Run**. It serves as the system's central orchestrator, handling all business logic.  
* **Authentication & Authorization**: **Firebase Authentication** is the gatekeeper for the entire system, managing user sign-up, login, and the issuance of secure JWTs.  
* **Report Generation Service (Python)**: A **Cloud Function** written in Python is triggered asynchronously by messages on a **Cloud Pub/Sub** topic. This service performs the heavy lifting of querying geospatial data sources, analyzing hazards, and generating the final PDF report.  
* **Persistence Layer**: **Google Cloud Firestore**, a serverless NoSQL database, stores all application data in separate collections for users, customers, property\_addresses, and report\_runs.  
* **Storage**: Final PDF reports are stored in **Google Cloud Storage**.  
* **Email Delivery**: **SendGrid** is integrated to handle the automated emailing of completed reports to customers.

## **Protocol Buffer Definitions**

Protocol Buffers (Protobufs) are used as the canonical data format throughout the system, ensuring type safety and consistency between the Go and Python services.

```
syntax = "proto3";

package nhdreport;

import "google/protobuf/timestamp.proto";

// ========== User ==========
message Permissions {
  bool can_create_customers = 1;
  bool can_generate_reports = 2;
  bool is_admin = 3;
}
message User {
  string user_id = 1; // Firebase Auth UID
  string full_name = 2;
  string email = 3;
  Permissions permissions = 4;
  google.protobuf.Timestamp created_at = 5;
}

// ========== Customer ==========
message Customer {
  string customer_id = 1;
  string full_name = 2;
  string email = 3;
  string company_name = 4;
  google.protobuf.Timestamp created_at = 5;
  string created_by_user_id = 6;
}

// ========== Property Address ==========
message PropertyAddress {
  string property_address_id = 1;
  message AddressDetails {
    string street_address = 1;
    string street_address_2 = 2;
    string city = 3;
    string state = 4;
    string zip_code = 5;
    string zip_plus_4 = 6;
  }
  AddressDetails address_details = 2;
  message Coordinates {
    double latitude = 1;
    double longitude = 2;
  }
  Coordinates coordinates = 3;
  string plus_code = 4;
  string google_place_id = 5;
}

// ========== Report Run ==========
message ReportRun {
  string report_run_id = 1;
  string customer_id = 2;
  // The ID of the internal user who created the report run on behalf of a customer.
  // This field is optional. If it's not set, it implies the customer
  // (identified by customer_id) created the report for themselves.
  string created_by_user_id = 3;
  string property_address_id = 4;
  google.protobuf.Timestamp created_at = 5;
  enum Status {
    STATUS_UNSPECIFIED = 0;
    PENDING = 1;
    PROCESSING = 2;
    COMPLETED = 3;
    FAILED = 4;
  }
  Status status = 6;
  message HazardResults {
    bool in_special_flood_hazard_area = 1;
    bool in_dam_inundation_area = 2;
    bool in_very_high_fire_hazard_severity_zone = 3;
    bool in_wildland_fire_area = 4;
    bool in_earthquake_fault_zone = 5;
    bool in_seismic_hazard_zone = 6;
  }
  HazardResults results = 7;
  string template_reference = 8;
  string final_pdf_storage_path = 9;
  message EmailDelivery {
    enum DeliveryStatus {
      STATUS_UNSPECIFIED = 0;
      SENT = 1;
      FAILED = 2;
    }
    DeliveryStatus status = 1;
    google.protobuf.Timestamp sent_at = 2;
    string email_template_reference = 3;
  }
  repeated EmailDelivery email_deliveries = 10;
  bool disable_automatic_email = 11;

  // Financials
  message ReportCost {
    double amount = 1;
    string currency = 2; // e.g., "USD"
    google.protobuf.Timestamp set_at = 3;
    string set_by_user_id = 4;
  }
  repeated ReportCost cost_history = 12; // Complete, auditable history of cost changes.

  message Payment {
    enum PaymentStatus {
      PAYMENT_STATUS_UNSPECIFIED = 0;
      OUTSTANDING = 1;
      PAID = 2;
      REFUNDED = 3;
    }
    PaymentStatus status = 1;
    double amount_paid = 2;
    string currency = 3;
    google.protobuf.Timestamp paid_at = 4;
    string payment_method = 5; // e.g., "Stripe", "Manual"
    string transaction_id = 6;
  }
  Payment payment_details = 13;
}
```

## **API Endpoints**

The Go Backend API will expose the following RESTful endpoints:

* **Users**  
  * POST /users/register: Creates a user profile in Firestore after successful Firebase Authentication sign-up.  
* **Customers**  
  * POST /customers: Creates a new customer record.  
  * GET /customers: Retrieves a list of all customers.  
* **Report Runs**  
  * POST /report-runs: Initiates a new report generation run.  
  * GET /report-runs: Retrieves a list of report runs with support for filtering (including by payment\_status), sorting, and pagination.  
  * POST /report-runs/{id}/resend-email: Triggers the resending of a completed report email.  
  * PUT /report-runs/{id}/cost: Sets or updates the cost for a specific report run. Appends a new entry to the cost\_history for auditing.  
  * POST /report-runs/{id}/payment: Records a payment against a specific report run.  
* **Financials**  
  * GET /financials/summary: Retrieves an aggregate summary of paid reports over a specified time frame.

## **Detailed System Workflow**

### **1\. User & Customer Management**

A user first signs up and logs in via the **Frontend**, which is managed by **Firebase Authentication**. Upon first login, the **Backend API (Go)** creates a corresponding User profile in the users collection in Firestore with default permissions. The authenticated user can then create Customer records via a dedicated API endpoint, which are also stored in Firestore and tagged with the user's ID for auditing.

### **2\. Report Generation Run**

1. An authenticated user selects a customer, enters a property address, and specifies email preferences.  
2. The **Backend API (Go)** receives the request. It first checks if a PropertyAddress record for this location already exists; if not, it creates one.  
3. It then creates a new ReportRun document in Firestore with a "PENDING" status.  
4. **Cost Assignment**: The API assigns an initial cost to the report by adding the first ReportCost entry to the cost\_history. The payment\_details are initialized with a status of "OUTSTANDING".  
5. The API then publishes a message containing the unique report\_run\_id to a **Pub/Sub** topic.  
6. The **Report Generation Service (Python Cloud Function)** is triggered, performs its analysis, and generates the PDF.  
7. If applicable, the service sends the report via **SendGrid**.  
8. Finally, the function updates the Firestore document status to "COMPLETED".

## **Financials**

This section describes the system for managing the cost and payment status of NHD reports.

### **1\. Data Models & Auditing**

* **Cost Management**: The cost of a report is stored in the cost\_history array within the ReportRun document. The *current* cost is always the last entry in this array. When a user with appropriate permissions updates the price via the PUT /report-runs/{id}/cost endpoint, a new ReportCost object is appended to the array. This preserves the full history of who changed the price, to what, and when, creating a complete and auditable record.  
* **Payment Tracking**: The payment\_details object within the ReportRun document tracks the financial status of the report. It is updated via the POST /report-runs/{id}/payment endpoint when a payment is recorded.

### **2\. Web Interface: Financial Reporting**

A dedicated "Financials" section in the frontend application will provide views for financial management and reporting.

* **Outstanding Reports View**:  
  * **Purpose**: To show all reports that have not yet been paid for.  
  * **Functionality**: This view will display a table of all report runs where payment\_details.status is "OUTSTANDING". The table will include columns for the customer name, property address, amount due, and the date the report was created. This view is powered by the GET /report-runs?payment\_status=OUTSTANDING API call.  
* **Paid Reports Summary View**:  
  * **Purpose**: To provide an overview of revenue and paid reports over a given period.  
  * **Functionality**: This view will feature a date range selector (e.g., "This Month," "Last Quarter," "Custom Range"). When a range is selected, it calls the GET /financials/summary endpoint. The interface will display:  
    * A summary card showing the **Total Revenue** for the selected period.  
    * A detailed table listing every report paid in that timeframe, including customer, property address, amount paid, and the date of payment.

## **Web Interface: Viewing and Filtering Reports**

A key feature of the application is a dashboard page that allows users to view and manage all generated reports.

The **Backend API** provides the GET /report-runs endpoint that allows for querying the report\_runs collection. The endpoint supports parameters for filtering by customer\_id, created\_by\_user\_id, and status, as well as sorting and pagination. To ensure these queries are performant, **composite indexes** must be configured in Firestore.

The **Frontend Interface** contains a set of UI controls (dropdowns, buttons) that allow the user to build their desired view. When a filter is applied, the frontend makes a new API call with the appropriate query parameters and updates the main results table with the new data. Each report will have a "Resend Email" button to trigger the resend workflow.

## **Geospatial Hazard Analysis**

This section details the data sources and the technical process used by the Python Report Generation Service to determine if a property lies within any of the six legally required hazard zones.

### **Data Sources**

All geospatial data is sourced directly from the authoritative state or federal agency responsible for mapping each specific hazard. The data is typically provided in standard formats like **Shapefile** or **GeoJSON**, which are then pre-processed and stored in a queryable format (e.g., PostGIS or a spatially-indexed flat file database) accessible by the Cloud Function.

1. **Special Flood Hazard Area**:  
   * **Agency**: Federal Emergency Management Agency (FEMA)  
   * **Data Format**: Shapefile (from the National Flood Hazard Layer)  
2. **Dam Inundation Area**:  
   * **Agency**: California Governor's Office of Emergency Services (CalOES)  
   * **Data Format**: Shapefile  
3. **Very High Fire Hazard Severity Zone**:  
   * **Agency**: California Department of Forestry and Fire Protection (CAL FIRE)  
   * **Data Format**: Shapefile  
4. **Wildland Fire Area** (State Responsibility Areas):  
   * **Agency**: California Department of Forestry and Fire Protection (CAL FIRE)  
   * **Data Format**: Shapefile  
5. **Earthquake Fault Zone**:  
   * **Agency**: California Geological Survey (CGS)  
   * **Data Format**: Shapefile  
6. **Seismic Hazard Zone** (Liquefaction & Landslide):  
   * **Agency**: California Geological Survey (CGS)  
   * **Data Format**: Shapefile

### **Point-in-Polygon Analysis Process**

The core of the hazard analysis is a **point-in-polygon (PIP)** test. This geometric operation determines whether a given point (the property's coordinates) is inside, outside, or on the boundary of a polygon (the hazard zone).

The Python Cloud Function executes the following steps for each of the six hazards:

1. **Retrieve Coordinates**: The function reads the property's latitude and longitude from the PropertyAddress document in Firestore.  
2. **Query Hazard Data**: For a specific hazard (e.g., Very High Fire Hazard Severity Zone), the function queries its pre-processed geospatial data source to find all hazard zone polygons that are geographically near the property's coordinates. This initial spatial index query is crucial for performance, as it avoids checking against every polygon in the state.  
3. **Perform PIP Test**: The function then iterates through the nearby hazard polygons. Using a standard geospatial library like **Shapely** or **GeoPandas**, it performs a PIP test to see if the property's coordinate point is contained within any of these polygons.  
4. **Record Result**: If the point falls within any polygon for that hazard type, the corresponding boolean flag in the HazardResults message (e.g., in\_very\_high\_fire\_hazard\_severity\_zone) is set to true.  
5. **Aggregate Results**: This process is repeated for all six hazard types. The final, aggregated HazardResults are then saved back to the ReportRun document in Firestore.

## **Testing Strategy**

To ensure the reliability, correctness, and robustness of the NHD service, a multi-layered testing strategy will be implemented for the Go backend. This strategy is composed of unit, integration, and end-to-end tests.

### **1\. Unit Tests (Go)**

* **Scope**: Unit tests focus on the smallest components of the application—individual functions and methods—in complete isolation from external dependencies like databases or other services.  
* **Tooling**:  
  * Go's standard testing package.  
  * testify/assert for fluent assertions.  
  * testify/mock for creating mock implementations of external dependencies (e.g., a mock Firestore client or Pub/Sub publisher).  
* **Examples**:  
  * **Request Validation**: Testing the logic within an HTTP handler that validates an incoming request body, ensuring it rejects invalid data (e.g., a missing email for a new customer) and returns the correct HTTP status code.  
  * **Business Logic**: Testing a function that creates a new ReportRun struct, ensuring that fields like status and created\_at are initialized correctly before being sent to the database.  
  * **Data Transformation**: Testing any utility functions that transform data from one format to another.

### **2\. Integration Tests (Go)**

* **Scope**: Integration tests verify the interaction between the Go backend API and its direct external dependencies, primarily Firestore and Pub/Sub. These tests ensure that database queries are correct and that messages are published as expected.  
* **Tooling**:  
  * Go's testing package.  
  * **Testcontainers**: To programmatically spin up and manage ephemeral, local instances of GCP service emulators (e.g., the Firestore emulator and Pub/Sub emulator) in Docker containers for each test run. This provides a high-fidelity testing environment without connecting to live GCP services.  
* **Examples**:  
  * **POST /report-runs Endpoint**: A test that calls the "create report run" endpoint and then connects directly to the test container's Firestore emulator to verify that a ReportRun document was created with the correct data and "PENDING" status. It would also subscribe to the test Pub/Sub emulator to confirm that a message with the new report\_run\_id was published to the correct topic.  
  * **GET /report-runs Endpoint**: Tests that first seed the Firestore emulator with a set of ReportRun documents and then call the "get report runs" endpoint with various filter and pagination parameters to assert that the API returns the correct subset of data.

### **3\. End-to-End (E2E) Tests**

* **Scope**: E2E tests validate the entire application workflow from the perspective of a user. They simulate real user interactions and cover the complete flow across all microservices: Frontend \-\> Go Backend API \-\> Pub/Sub \-\> Python Cloud Function \-\> Firestore.  
* **Tooling**:  
  * A browser automation framework like **Cypress** or **Playwright**.  
  * A dedicated testing environment deployed on GCP that mirrors the production setup.  
* **Examples**:  
  * **Full Report Generation Flow**:  
    1. The E2E test script automates a browser to log into the web application.  
    2. It navigates to the report creation page, fills in a customer and a specific property address for which a "golden" record exists, and submits the form.  
    3. The script then polls the GET /report-runs API endpoint for the specific report, waiting for its status to change from "PENDING" to "PROCESSING" and finally to "COMPLETED".  
    4. Once the status is "COMPLETED", the script inspects the JSON response for the completed report run.  
    5. It extracts the results object (containing the boolean hazard flags) from the response.  
    6. This results object is then compared against a pre-defined "golden" data file (e.g., a JSON file) that contains the known, correct hazard values for that specific property address.  
    7. The test passes only if the generated HazardResults data perfectly matches the golden record, ensuring the accuracy of the entire geospatial analysis pipeline without the need for brittle PDF parsing.

## **Monitoring, Logging, and Alerting**

To ensure operational excellence, system health, and rapid incident response, the application will leverage the Google Cloud Observability suite (formerly Stackdriver).

### **1\. Metrics Collection (Cloud Monitoring)**

**Cloud Monitoring** will be used to automatically collect, visualize, and dashboard key metrics from all system components. This provides a real-time view of service performance and resource utilization.

* **Backend API (Cloud Run)**:  
  * **Request Latency** (especially p95 and p99) to detect performance degradation.  
  * **Request Count & Error Rate** (4xx and 5xx) to monitor API health and usage patterns.  
  * **Container CPU and Memory Utilization** to inform scaling policies.  
* **Report Generation Service (Cloud Functions)**:  
  * **Execution Count & Duration** to track processing times and identify bottlenecks.  
  * **Memory Usage** to ensure functions are correctly provisioned.  
  * **Error Rate** (including crashes and timeouts).  
* **Firestore**:  
  * **Read/Write/Delete Operations** to monitor database load.  
  * **Active Connections** to track usage.  
* **Pub/Sub**:  
  * **Publish/Subscribe Request Counts** to ensure messages are flowing.  
  * **Unacknowledged Message Count** to detect issues with the Python consumer function.

### **2\. Centralized Logging (Cloud Logging)**

All services will be configured to stream logs to **Cloud Logging**, providing a centralized and searchable repository for all application and system logs.

* **Structured Logging**: Both the Go and Python services will output logs in a structured **JSON format**. This is critical as it allows for powerful filtering and analysis. For example, logs will include standard fields like severity, timestamp, and service\_name, as well as context-specific fields like report\_run\_id or user\_id.  
* **Log Correlation**: By including a common trace ID across services, logs can be correlated to trace a single user request as it flows through the entire system, from the initial API call to the final PDF generation.

### **3\. Alerting Strategy**

A proactive alerting system will be built within **Cloud Monitoring** to notify the engineering team of potential issues before they impact users.

* **Metric-Based Alerts**: Alerting policies will be created based on thresholds for the metrics collected above.  
  * *Example 1*: An alert is triggered if the p99 latency for the POST /report-runs endpoint exceeds 2 seconds for more than 5 minutes.  
  * *Example 2*: An alert is triggered if the error rate of the Python Cloud Function exceeds 1% over a 10-minute window.  
* **Log-Based Alerts**: Using the native log parsing and alerting features of Cloud Logging, alerts will be configured to trigger when specific patterns appear in the logs. This is used for catching application-level errors that don't always manifest as a simple metric spike.  
  * *Example 1*: An alert is triggered immediately if a log entry with severity: "ERROR" containing a stack trace is detected in the Python Report Generation Service.  
  * *Example 2*: An alert is triggered if the Go API logs a "failed to publish to Pub/Sub" error message.  
* **Notification Channels**: Alerts will be routed to pre-configured notification channels, including:  
  * **Email** for low-priority warnings.  
  * **Slack** for medium-priority incidents that require team awareness.  
  * **PagerDuty** for high-priority, critical incidents that require immediate, on-call response.
