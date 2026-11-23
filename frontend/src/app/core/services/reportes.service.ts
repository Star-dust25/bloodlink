import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { API_CONFIG } from '../config/api.config';
import { DashboardMetricas, TrazabilidadUnidadReporte } from '../models/reporte.model';

@Injectable({ providedIn: 'root' })
export class ReportesService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${API_CONFIG.apiV1}/reportes`;

  obtenerDashboard(): Observable<DashboardMetricas> {
    return this.http.get<DashboardMetricas>(`${this.baseUrl}/dashboard`);
  }

  obtenerTrazabilidadJSON(unidadId: string): Observable<TrazabilidadUnidadReporte> {
    const params = new HttpParams().set('format', 'json');
    return this.http.get<TrazabilidadUnidadReporte>(
      `${this.baseUrl}/trazabilidad/${encodeURIComponent(unidadId)}`,
      { params }
    );
  }

  descargarTrazabilidadArchivo(unidadId: string, format: 'pdf' | 'csv'): Observable<Blob> {
    const params = new HttpParams().set('format', format);
    return this.http.get(`${this.baseUrl}/trazabilidad/${encodeURIComponent(unidadId)}`, {
      params,
      responseType: 'blob',
    });
  }
}
