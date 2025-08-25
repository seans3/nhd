const API_BASE_URL = 'http://localhost:8080';

/**
 * Fetches a list of all report runs from the backend.
 * @returns {Promise<Array>} A promise that resolves to an array of report runs.
 */
export const getReportRuns = async () => {
  try {
    const response = await fetch(`${API_BASE_URL}/report-runs`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return await response.json();
  } catch (error) {
    console.error('Error fetching report runs:', error);
    return []; // Return empty array on error
  }
};

/**
 * Creates a new customer.
 * @param {Object} customerData - The customer data to create.
 * @returns {Promise<Object>} A promise that resolves to the newly created customer data.
 */
export const createCustomer = async (customerData) => {
  const response = await fetch(`${API_BASE_URL}/customers`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(customerData),
  });
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return await response.json();
};

/**
 * Creates a new report run.
 * @param {Object} reportData - The report run data to create.
 * @returns {Promise<Object>} A promise that resolves to the newly created report run data.
 */
export const createReportRun = async (reportData) => {
  const response = await fetch(`${API_BASE_URL}/report-runs`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(reportData),
  });
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return await response.json();
};


/**
 * Fetches the financials summary from the backend.
 * @returns {Promise<Object>} A promise that resolves to the financial summary data.
 */
export const getFinancialsSummary = async () => {
  try {
    const response = await fetch(`${API_BASE_URL}/financials/summary`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();
    // Ensure paid_reports is always an array
    if (!data.paid_reports) {
      data.paid_reports = [];
    }
    return data;
  } catch (error) {
    console.error('Error fetching financials summary:', error);
    // Return a default structure on error to prevent the UI from breaking
    return {
      total_revenue: 0.00,
      paid_reports: [],
    };
  }
};
