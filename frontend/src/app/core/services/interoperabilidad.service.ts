import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { API_CONFIG } from '../config/api.config';
import { HealthStatus, UDDIDirectorio } from '../models/interoperabilidad.model';

@Injectable({ providedIn: 'root' })
export class InteroperabilidadService {
  private readonly http = inject(HttpClient);

  verificarHealth(): Observable<HealthStatus> {
    return this.http.get<HealthStatus>(API_CONFIG.health);
  }

  obtenerDirectorioUDDI(): Observable<UDDIDirectorio> {
    return this.http.get<UDDIDirectorio>(`${API_CONFIG.apiV1}/uddi/servicios`);
  }

  obtenerContratoWSDL(): Observable<string> {
    return this.http.get(`${API_CONFIG.soapV1}?wsdl`, { responseType: 'text' });
  }
}
