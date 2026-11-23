export const API_CONFIG = {
  apiV1: '/api/v1',
  soapV1: '/soap/v1/interoperabilidad',
  health: '/health',
} as const;

export interface BancoCatalogoItem {
  id: string;
  nombre: string;
  tipoInstitucion: 'MINSA' | 'EsSalud' | 'Privada';
  coordenadas: string;
  jurisdiccion: string;
  esLegacySoap: boolean;
}

export const BANCOS_PIURA_CATALOGO: readonly BancoCatalogoItem[] = [
  {
    id: 'HOSP-MINSA-01',
    nombre: 'Hospital Nacional Santa Rosa - Piura',
    tipoInstitucion: 'MINSA',
    coordenadas: '-5.1945,-80.6328',
    jurisdiccion: 'Piura Oeste',
    esLegacySoap: false,
  },
  {
    id: 'HOSP-ESSALUD-04',
    nombre: 'Hospital III José Cayetano Heredia - EsSalud',
    tipoInstitucion: 'EsSalud',
    coordenadas: '-5.1821,-80.6190',
    jurisdiccion: 'Castilla',
    esLegacySoap: true,
  },
  {
    id: 'HOSP-MINSA-02',
    nombre: 'Hospital de Apoyo II-2 Sullana',
    tipoInstitucion: 'MINSA',
    coordenadas: '-4.9039,-80.6853',
    jurisdiccion: 'Sullana',
    esLegacySoap: false,
  },
  {
    id: 'CLIN-PRIV-01',
    nombre: 'Clínica Miraflores - Piura',
    tipoInstitucion: 'Privada',
    coordenadas: '-5.1880,-80.6250',
    jurisdiccion: 'Castilla',
    esLegacySoap: false,
  },
] as const;

export const GRUPOS_SANGUINEOS = ['O', 'A', 'B', 'AB'] as const;
export const FACTORES_RH = ['+', '-'] as const;
