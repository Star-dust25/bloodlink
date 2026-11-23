import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { API_CONFIG } from '../config/api.config';
import {
  CancelarEmergenciaRequest,
  CrearEmergenciaRequest,
  EmergenciaMatchingResponseDTO,
} from '../models/emergencia.model';

@Injectable({ providedIn: 'root' })
export class EmergenciasService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${API_CONFIG.apiV1}/emergencias`;

  crearSolicitudYMatching(payload: CrearEmergenciaRequest): Observable<EmergenciaMatchingResponseDTO> {
    return this.http.post<EmergenciaMatchingResponseDTO>(this.baseUrl, payload);
  }

  listarSolicitudes(estado?: string): Observable<EmergenciaMatchingResponseDTO[]> {
    let params = new HttpParams();
    if (estado) params = params.set('estado', estado);
    return this.http.get<EmergenciaMatchingResponseDTO[]>(this.baseUrl, { params });
  }

  obtenerSolicitud(solicitudId: string): Observable<EmergenciaMatchingResponseDTO> {
    return this.http.get<EmergenciaMatchingResponseDTO>(
      `${this.baseUrl}/${encodeURIComponent(solicitudId)}`
    );
  }

  cancelarSolicitud(
    solicitudId: string,
    payload: CancelarEmergenciaRequest
  ): Observable<EmergenciaMatchingResponseDTO> {
    return this.http.patch<EmergenciaMatchingResponseDTO>(
      `${this.baseUrl}/${encodeURIComponent(solicitudId)}/cancelar`,
      payload
    );
  }

  confirmarTransfusion(solicitudId: string): Observable<EmergenciaMatchingResponseDTO> {
    return this.http.put<EmergenciaMatchingResponseDTO>(
      `${this.baseUrl}/${encodeURIComponent(solicitudId)}/confirmar`,
      {}
    );
  }
}
