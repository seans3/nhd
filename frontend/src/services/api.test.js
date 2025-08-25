import { getReportRuns, createCustomer } from './api';

describe('API Service', () => {
  beforeEach(() => {
    fetch.resetMocks();
  });

  test('getReportRuns fetches and returns data', async () => {
    const mockData = [{ report_run_id: 'run1' }];
    fetch.mockResponseOnce(JSON.stringify(mockData));

    const data = await getReportRuns();

    expect(fetch).toHaveBeenCalledWith('http://localhost:8080/report-runs');
    expect(data).toEqual(mockData);
  });

  test('createCustomer sends the correct data and returns the response', async () => {
    const mockCustomer = { full_name: 'John Doe', email: 'john@example.com' };
    const mockResponse = { customer_id: 'cust_123' };
    fetch.mockResponseOnce(JSON.stringify(mockResponse));

    const data = await createCustomer(mockCustomer);

    expect(fetch).toHaveBeenCalledWith('http://localhost:8080/customers', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(mockCustomer),
    });
    expect(data).toEqual(mockResponse);
  });

  test('getReportRuns handles API errors gracefully', async () => {
    fetch.mockReject(new Error('API is down'));

    const data = await getReportRuns();

    expect(data).toEqual([]); // Should return an empty array on error
  });
});
