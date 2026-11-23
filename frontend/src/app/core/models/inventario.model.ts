export type TipoSangre = 'O' | 'A' | 'B' | 'AB';
export type FactorRh = '+' | '-';
export type EstadoUnidad = 'Disponible' | 'Reservada' | 'Transfundida' | 'Caducada';
export type AlertaCaducidad = 'VERDE' | 'AMARILLO' | 'ROJO';

export interface BancoStockResumenDTO {
  bancoId: string;
  nombre: string;
  tipoInstitucion: string;
  coordenadas: string;
  jurisdiccion: string;
  esLegacySoap: boolean;
  unidadesDisponibles: number;
  proximasACaducar7Dias: number;
}

export interface StockRedResponseDTO {
  tipoSangre: string;
  factorRh: string;
  stockGlobalDisponible: number;
  nivelAbastecimiento: 'ESTABLE' | 'CRITICO';
  bancos: BancoStockResumenDTO[];
  consultadoEn: string;
}

export interface RegistrarUnidadRequest {
  bancoId: string;
  tipoSangre: TipoSangre;
  factorRh: FactorRh;
  volumenMl: number;
  fechaExtraccion: string;
  origen: string;
}

export interface UnidadSangreResponseDTO {
  unidadId: string;
  codigoBarras: string;
  bancoId: string;
  tipoSangre: TipoSangre;
  factorRh: FactorRh;
  volumenMl: number;
  origen: string;
  fechaExtraccion: string;
  fechaCaducidad: string;
  estado: EstadoUnidad;
  alertaCaducidad: AlertaCaducidad;
  diasRestantes: number;
}

export interface ActualizarEstadoRequest {
  nuevoEstado: EstadoUnidad;
  motivo?: string;
}
