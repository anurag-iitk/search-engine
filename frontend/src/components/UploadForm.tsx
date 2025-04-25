import { useState } from 'react';
import { uploadFile } from '../api';

const UploadForm = () => {
  const [file, setFile] = useState<File | null>(null);

  const handleUpload = async () => {
    if (!file) return;
    try {
      await uploadFile(file);
      alert('File uploaded successfully!');
    } catch (err) {
      alert('Upload failed.');
    }
  };

  return (
    <div className="upload-card">
  <div className="upload-card-content">
    <input
      type="file"
      accept=".parquet"
      onChange={(e) => setFile(e.target.files?.[0] || null)}
      className="upload-input"
    />
    <button onClick={handleUpload} className="upload-button">
      Upload
    </button>
  </div>
</div>

  );
};

export default UploadForm;
