const API_BASE_URL = 'http://localhost:8080';

/**
 * Fetches the financials summary from the backend.
 * NOTE: This will not work until the backend endpoint is implemented.
 * @returns {Promise<Object>} A promise that resolves to the financial summary data.
 */
export const getFinancialsSummary = async () => {
  try {
    // This endpoint is currently a stub on the backend.
    // When implemented, it should return data like:
    // { total_revenue: 1234.56, paid_reports: [...] }
    const response = await fetch(`${API_BASE_URL}/financials/summary`);
    if (!response.ok) {
      // The real endpoint returns 501 Not Implemented right now.
      // We'll return mock data to allow for UI development.
      console.warn('Financials summary endpoint is not implemented. Returning mock data.');
      return {
        total_revenue: 5432.10,
        paid_reports: [
          { customer_name: 'Alice', property_address: '123 Main St', amount_paid: 50.00, paid_at: '2025-08-24' },
          { customer_name: 'Bob', property_address: '456 Oak Ave', amount_paid: 75.50, paid_at: '2025-08-23' },
        ],
      };
    }
    return await response.json();
  } catch (error) {
    console.error('Error fetching financials summary:', error);
    // Return mock data on error to prevent UI from breaking
    return {
      total_revenue: 5432.10,
      paid_reports: [
        { customer_name: 'Alice', property_address: '123 Main St', amount_paid: 50.00, paid_at: '2025-08-24' },
        { customer_name: 'Bob', property_address: '456 Oak Ave', amount_paid: 75.50, paid_at: '2025-08-23' },
      ],
    };
  }
};
