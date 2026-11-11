import { request } from '../request';

/** Check connectivity with the Go backend. */
export function fetchHello() {
  return request<string, 'text'>({ url: '/hello', responseType: 'text' });
}