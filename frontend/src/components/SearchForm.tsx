import { useState } from 'react';
import { useAppDispatch } from '../app/store';
import { performSearch } from '../features/search/searchSlice';

// All available searchable fields based on Parquet schema
const SEARCHABLE_FIELDS = [
  'Namespace',
  'MsgId',
  'PartitionId',
  'Timestamp',
  'Hostname',
  'Priority',
  'Facility',
  'FacilityString',
  'Severity',
  'SeverityString',
  'AppName',
  'ProcId',
  'Message',
  'MessageRaw',
  'StructuredData',
  'Tag',
  'Sender',
  'Groupings',
  'Event',
  'EventId',
  'NanoTimeStamp',
];

const SearchForm = () => {
  const dispatch = useAppDispatch();
  const [query, setQuery] = useState('');
  const [field, setField] = useState('Namespace'); // Default field

  const handleSearch = () => {
    if (!query.trim()) return;
    dispatch(performSearch({ query, field }));
  };

  return (
    <div className="search-form-container">
      <div className="search-form">
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Enter search term"
          className="search-input"
        />

        <select
          value={field}
          onChange={(e) => setField(e.target.value)}
          className="search-select"
        >
          {SEARCHABLE_FIELDS.map((f) => (
            <option key={f} value={f}>
              {f}
            </option>
          ))}
        </select>

        <button onClick={handleSearch} className="search-button">
          Search
        </button>
      </div>
    </div>
  );
};

export default SearchForm;
