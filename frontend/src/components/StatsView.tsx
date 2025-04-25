import { useEffect, useState } from 'react';
import { getStats } from '../api';

const StatsView = () => {
  const [stats, setStats] = useState<any>(null);

  useEffect(() => {
    getStats().then((res) => setStats(res.data));
  }, []);

  return (
    <div className="stats-container">
      <h2 className="stats-title">📊 Server Stats</h2>
      {stats ? (
        <div className="stats-card">
          <div className="stat-item">
            <span className="stat-label">Total Documents</span>
            <span className="stat-value">{stats.totalDocuments}</span>
          </div>
          <div className="stat-item">
            <span className="stat-label">Search Count</span>
            <span className="stat-value">{stats.searchCount}</span>
          </div>
          <div className="stat-item">
            <span className="stat-label">Average Search Time</span>
            <span className="stat-value">{stats.avgSearchTime}</span>
          </div>
        </div>
      ) : (
        <p className="loading-text">Loading stats...</p>
      )}
    </div>
  );
};

export default StatsView;
