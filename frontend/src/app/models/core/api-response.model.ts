export interface ApiResponse<T> {
  data: T;
  count: string | number;
}

export interface DataEnvelope<T> {
  data: T;
}
