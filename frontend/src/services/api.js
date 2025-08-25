const API_BASE_URL = 'http://localhost:8080';

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
