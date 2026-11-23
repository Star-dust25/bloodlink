import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { API_CONFIG } from '../config/api.config';
import { Notificacion, VerificacionCaducidadResponse } from '../models/notificacion.model';

@Injectable({ providedIn: 'root' })
export class NotificacionesService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${API_CONFIG.apiV1}/notificaciones`;

  listarNotificaciones(canal?: string, limit = 50): Observable<Notificacion[]> {
    let params = new HttpParams().set('limit', String(limit));
    if (canal) params = params.set('canal', canal);
    return this.http.get<Notificacion[]>(this.baseUrl, { params });
  }

  ejecutarVerificacionCaducidad(): Observable<VerificacionCaducidadResponse> {
    return this.http.post<VerificacionCaducidadResponse>(`${this.baseUrl}/verificar-caducidad`, {});
  }
}
