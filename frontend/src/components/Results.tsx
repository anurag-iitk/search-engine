import { useSelector } from 'react-redux';
import { RootState } from '../app/store';

const Results = () => {
  const { data, loading, error } = useSelector((state: RootState) => state.search);

  if (loading) return <p style={{ color: '#2563eb' }}>🔄 Loading...</p>;
  if (error) return <p style={{ color: '#dc2626' }}>❌ Error: {error}</p>;
  if (!data || data.results.length === 0) return <p>No results found.</p>;

  return (
    <div style={{ marginTop: '2rem' }}>
      <h3 style={{ marginBottom: '1rem' }}>
        Results ({data.results.length}) — Time: {data.timeTaken}
      </h3>

      <div style={{ display: 'grid', gap: '1rem' }}>
        {data.results.map((result, index) => {
          const record = result.Document?.Record;
          const score = result.Score;

          return (
            <div
              key={index}
              style={{
                padding: '1rem',
                backgroundColor: '#f9f9f9',
                border: '1px solid #ddd',
                borderRadius: '8px',
              }}
            >
              <p>
                <strong>Score:</strong> {score?.toFixed(6)}
              </p>

              {record ? (
                <div style={{ fontFamily: 'monospace', fontSize: '0.9rem' }}>
                  {Object.entries(record).map(([key, value]) => (
                    <div key={key}>
                      <strong>{key}:</strong> {String(value)}
                    </div>
                  ))}
                </div>
              ) : (
                <p style={{ color: '#999' }}>No record data found.</p>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default Results;
