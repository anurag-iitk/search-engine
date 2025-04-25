import axios from 'axios';

const API = axios.create({
  baseURL: 'http://localhost:8080',
});

export const getStats = () => API.get('/stats');

export const uploadFile = (file: File) => {
  const formData = new FormData();
  formData.append('files', file);
  return API.post('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
};
