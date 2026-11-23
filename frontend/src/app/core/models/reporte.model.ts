export interface StockPorGrupoKPI {
  grupoSanguineo: string;
  tipoSangre: string;
  factorRh: string;
  unidadesDisponibles: number;
  unidadesReservadas: number;
}

export interface EmergenciaEstadoKPI {
  estado: string;
  cantidad: number;
}

export interface EventoAuditItem {
  id: number;
  entidad: string;
  entidadId: string;
  accion: string;
  estadoAnterior: string;
  estadoNuevo: string;
  usuarioId: string;
  detalle: string;
  timestamp: string;
}

export interface NotificacionResumenItem {
  id: number;
  evento: string;
  canal: string;
  destinatario: string;
  mensaje: string;
  timestamp: string;
}

export interface DashboardMetricas {
  totalBancosRed: number;
  totalUnidadesDisponibles: number;
  unidadesProximasACaducar7d: number;
  unidadesAlertaRoja3d: number;
  totalEmergencias: number;
  emergenciasActivas: number;
  emergenciasAtendidas: number;
  tiempoPromedioMatchingMs: number;
  donantesHabilitados: number;
  stockPorTipoSangre: StockPorGrupoKPI[];
  emergenciasPorEstado: EmergenciaEstadoKPI[];
  ultimasNotificaciones: NotificacionResumenItem[];
  generadoEn: string;
}

export interface TrazabilidadUnidadReporte {
  unidadId: string;
  codigoBarras: string;
  bancoId: string;
  bancoNombre: string;
  tipoSangre: string;
  factorRh: string;
  volumenMl: number;
  origen: string;
  fechaExtraccion: string;
  fechaCaducidad: string;
  estadoActual: string;
  historial: EventoAuditItem[];
  generadoEn: string;
}
