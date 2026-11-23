import { HttpErrorResponse } from '@angular/common/http';
import { ErrorResponse } from '../models/error.model';

export interface ExtractedError {
  status: number;
  codigo: string;
  mensaje: string;
  detalle: string;
}

export function extractApiError(err: unknown): ExtractedError {
  if (err instanceof HttpErrorResponse) {
    const body = err.error as Partial<ErrorResponse> | undefined;
    if (body && typeof body === 'object' && body.error) {
      return {
        status: err.status,
        codigo: body.codigo || `HTTP_${err.status}`,
        mensaje: body.error,
        detalle: body.detalle || '',
      };
    }
    if (err.status === 0) {
      return {
        status: 0,
        codigo: 'NETWORK_UNREACHABLE',
        mensaje: 'No fue posible establecer conexión con el servidor BloodLink.',
        detalle: 'Verifique que el backend se encuentre activo en el puerto 8080.',
      };
    }
    return {
      status: err.status,
      codigo: `HTTP_${err.status}`,
      mensaje: err.message || 'Ocurrió un error al procesar la solicitud.',
      detalle: '',
    };
  }
  return {
    status: 500,
    codigo: 'UNEXPECTED_ERROR',
    mensaje: 'Error inesperado en la aplicación.',
    detalle: err instanceof Error ? err.message : '',
  };
}
