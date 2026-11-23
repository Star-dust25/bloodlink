import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { API_CONFIG } from '../config/api.config';
import {
  CiudadanoReniec,
  DonanteResponseDTO,
  RegistrarDonanteRequest,
} from '../models/donante.model';

@Injectable({ providedIn: 'root' })
export class DonantesService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${API_CONFIG.apiV1}/donantes`;

  validarDniReniec(dni: string): Observable<CiudadanoReniec> {
    return this.http.post<CiudadanoReniec>(`${this.baseUrl}/validar-dni`, { dni });
  }

  registrarDonante(payload: RegistrarDonanteRequest): Observable<DonanteResponseDTO> {
    return this.http.post<DonanteResponseDTO>(this.baseUrl, payload);
  }

  listarDonantes(tipoSangre?: string, factorRh?: string): Observable<DonanteResponseDTO[]> {
    let params = new HttpParams();
    if (tipoSangre) params = params.set('tipoSangre', tipoSangre);
    if (factorRh) params = params.set('factorRh', factorRh);
    return this.http.get<DonanteResponseDTO[]>(this.baseUrl, { params });
  }

  darDeBaja(donanteId: string): Observable<{ mensaje: string; donanteId: string }> {
    return this.http.delete<{ mensaje: string; donanteId: string }>(
      `${this.baseUrl}/${encodeURIComponent(donanteId)}`
    );
  }
}
