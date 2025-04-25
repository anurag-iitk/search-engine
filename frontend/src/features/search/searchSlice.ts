import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import axios from 'axios';
import { SearchQuery, SearchResult } from './types';

interface SearchState {
  loading: boolean;
  error: string | null;
  data: SearchResult | null;
}

const initialState: SearchState = {
  loading: false,
  error: null,
  data: null,
};

export const performSearch = createAsyncThunk(
  'search/performSearch',
  async (payload: SearchQuery) => {
    const res = await axios.post<SearchResult>('http://localhost:8080/search', payload);
    return res.data;
  }
);

const searchSlice = createSlice({
  name: 'search',
  initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder
      .addCase(performSearch.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(performSearch.fulfilled, (state, action) => {
        state.loading = false;
        state.data = action.payload;
      })
      .addCase(performSearch.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Something went wrong';
      });
  },
});

export default searchSlice.reducer;
