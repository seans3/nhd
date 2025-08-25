import React, { useState, useEffect } from 'react';
import { getFinancialsSummary } from '../services/api';

function Financials() {
  const [summary, setSummary] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchSummary = async () => {
      try {
        const data = await getFinancialsSummary();
        setSummary(data);
      } catch (error) {
        console.error("Failed to fetch financials summary:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchSummary();
  }, []);

  if (loading) {
    return <div>Loading...</div>;
  }

  if (!summary) {
    return <div>Failed to load financial data.</div>;
  }

  return (
    <div style={{ padding: '20px' }}>
      <h1>Financials</h1>
      
      <div style={{ marginBottom: '20px', padding: '10px', border: '1px solid #ccc' }}>
        <h2>Total Revenue</h2>
        <p style={{ fontSize: '24px', fontWeight: 'bold' }}>
          ${summary.total_revenue.toFixed(2)}
        </p>
      </div>

      <h2>Paid Reports</h2>
      <table style={{ width: '100%', borderCollapse: 'collapse' }}>
        <thead>
          <tr style={{ borderBottom: '1px solid #ccc' }}>
            <th style={{ textAlign: 'left', padding: '8px' }}>Customer</th>
            <th style={{ textAlign: 'left', padding: '8px' }}>Property Address</th>
            <th style={{ textAlign: 'left', padding: '8px' }}>Amount Paid</th>
            <th style={{ textAlign: 'left', padding: '8px' }}>Date</th>
          </tr>
        </thead>
        <tbody>
          {summary.paid_reports.map((report, index) => (
            <tr key={index} style={{ borderBottom: '1px solid #eee' }}>
              <td style={{ padding: '8px' }}>{report.customer_name}</td>
              <td style={{ padding: '8px' }}>{report.property_address}</td>
              <td style={{ padding: '8px' }}>${report.amount_paid.toFixed(2)}</td>
              <td style={{ padding: '8px' }}>{report.paid_at}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export default Financials;
