import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { API_CONFIG } from '../config/api.config';
import {
  ActualizarEstadoRequest,
  RegistrarUnidadRequest,
  StockRedResponseDTO,
  UnidadSangreResponseDTO,
} from '../models/inventario.model';

export interface FiltroInventarioQuery {
  tipoSangre?: string;
  factorRh?: string;
  estado?: string;
  hospitalId?: string;
}

@Injectable({ providedIn: 'root' })
export class InventarioService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${API_CONFIG.apiV1}/inventario`;

  consultarStockRed(filtro: FiltroInventarioQuery = {}): Observable<StockRedResponseDTO> {
    let params = new HttpParams();
    if (filtro.tipoSangre) params = params.set('tipoSangre', filtro.tipoSangre);
    if (filtro.factorRh) params = params.set('factorRh', filtro.factorRh);
    if (filtro.estado) params = params.set('estado', filtro.estado);
    if (filtro.hospitalId) params = params.set('hospitalId', filtro.hospitalId);

    return this.http.get<StockRedResponseDTO>(`${this.baseUrl}/stock`, { params });
  }

  listarUnidades(filtro: FiltroInventarioQuery = {}): Observable<UnidadSangreResponseDTO[]> {
    let params = new HttpParams();
    if (filtro.tipoSangre) params = params.set('tipoSangre', filtro.tipoSangre);
    if (filtro.factorRh) params = params.set('factorRh', filtro.factorRh);
    if (filtro.estado) params = params.set('estado', filtro.estado);
    if (filtro.hospitalId) params = params.set('hospitalId', filtro.hospitalId);

    return this.http.get<UnidadSangreResponseDTO[]>(`${this.baseUrl}/unidades`, { params });
  }

  obtenerUnidad(unidadId: string): Observable<UnidadSangreResponseDTO> {
    return this.http.get<UnidadSangreResponseDTO>(`${this.baseUrl}/unidades/${encodeURIComponent(unidadId)}`);
  }

  registrarUnidad(payload: RegistrarUnidadRequest): Observable<UnidadSangreResponseDTO> {
    return this.http.post<UnidadSangreResponseDTO>(`${this.baseUrl}/unidades`, payload);
  }

  actualizarEstado(unidadId: string, payload: ActualizarEstadoRequest): Observable<UnidadSangreResponseDTO> {
    return this.http.patch<UnidadSangreResponseDTO>(
      `${this.baseUrl}/unidades/${encodeURIComponent(unidadId)}/estado`,
      payload
    );
  }
}
