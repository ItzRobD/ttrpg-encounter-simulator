import { HttpInterceptorFn, HttpResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { map } from 'rxjs';
import { MapperService } from '../services/mapper.service';

/**
 * Interceptor that automatically maps snake_case backend responses to camelCase
 * and camelCase frontend requests to snake_case for the backend.
 */
export const mappingInterceptor: HttpInterceptorFn = (req, next) => {
  const mapperService = inject(MapperService);

  // 1. Handle Request (Outbound)
  let modifiedReq = req;
  if (req.body && (req.method === 'POST' || req.method === 'PUT' || req.method === 'PATCH')) {
    const snakeBody = mapperService.serializeKeys(req.body);
    modifiedReq = req.clone({ body: snakeBody });
  }

  // 2. Handle Response (Inbound)
  return next(modifiedReq).pipe(
    map(event => {
      if (event instanceof HttpResponse && event.body) {
        // Map the response body to camelCase
        const camelBody = mapperService.mapKeys(event.body);
        return event.clone({ body: camelBody });
      }
      return event;
    })
  );
};
