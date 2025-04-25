export interface SearchQuery {
    query: string;
    field: string;
  }
  
  export interface SearchResult {
    results: any[];
    matchCount: number;
    timeTaken: string;
  }
  