export interface HealthStatus {
  status: 'UP' | 'DEGRADED';
  service: string;
  postgres: boolean;
  eventBus: boolean;
  timestamp: string;
}

export interface UDDIPaginaBlanca {
  organizacion: string;
  entidad: string;
  contacto: string;
  ruc: string;
  jurisdiccion: string;
}

export interface UDDIPaginaAmarilla {
  categoria: string;
  codigoUnspsc: string;
  estandar: string;
}

export interface UDDIPaginaVerde {
  operacion: string;
  protocolo: string;
  bindingUrl: string;
  soapAction: string;
  descripcion: string;
}

export interface UDDIDirectorio {
  servicio: string;
  version: string;
  namespace: string;
  wsdlEndpoint: string;
  paginasBlancas: UDDIPaginaBlanca;
  paginasAmarillas: UDDIPaginaAmarilla[];
  paginasVerdes: UDDIPaginaVerde[];
}
