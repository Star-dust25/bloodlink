import { FactorRh, TipoSangre } from './inventario.model';

export type EstadoSolicitud =
  | 'Registrada'
  | 'Buscando Stock'
  | 'Reservada'
  | 'En Espera de Donantes'
  | 'Atendida'
  | 'Cancelada';

export type OrigenAsignacion = 'LOCAL' | 'RED_EXTERNA' | 'NINGUNO';
export type UrgenciaNivel = 'ALTA_VITAL' | 'MEDIA';

export interface CrearEmergenciaRequest {
  hospitalOrigenId: string;
  coordenadasOrigen: string;
  medicoId: number;
  tipoSangreReq: TipoSangre;
  factorRhReq: FactorRh;
  cantidadReq: number;
  urgencia: UrgenciaNivel;
  simularFalloMapas?: boolean;
}

export interface BancoAsignadoDTO {
  bancoId: string;
  nombre: string;
  distanciaKm: number;
  tiempoEstimadoMin: number;
}

export interface UnidadAsignada {
  unidadId: string;
  codigoBarras: string;
}

export interface EmergenciaMatchingResponseDTO {
  solicitudId: string;
  estadoSolicitud: EstadoSolicitud;
  origenAsignacion: OrigenAsignacion;
  hospitalOrigenId?: string;
  tipoSangreReq?: TipoSangre;
  factorRhReq?: FactorRh;
  cantidadReq?: number;
  urgencia?: UrgenciaNivel;
  bancoAsignado?: BancoAsignadoDTO;
  unidadesAsignadas: UnidadAsignada[];
  tiempoRespuestaMatchingMs: number;
  fallbackJurisdiccionUsado: boolean;
  motivoCancelacion?: string;
  fechaHora: string;
}

export interface EmergenciaSinStockConflictDTO {
  solicitudId: string;
  estadoSolicitud: EstadoSolicitud;
  mensaje: string;
  sugerirConvocatoriaDonantes: boolean;
  tiempoRespuestaMatchingMs: number;
  fechaHora: string;
}

export interface CancelarEmergenciaRequest {
  motivo: string;
}
