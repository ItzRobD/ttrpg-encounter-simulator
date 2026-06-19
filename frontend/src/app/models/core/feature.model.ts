export interface Feature {
  id: string;
  name: string;
  description: string;
  hooks?: Record<string, boolean>;
  data?: Record<string, any>;
}
