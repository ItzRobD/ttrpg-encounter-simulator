export const environment = {
  // Base/default environment (ng serve / builds without a config). Production
  // builds replace this with environment.prod.ts (see angular.json).
  name: 'default',
  apiUrl: 'http://localhost:8080/api/v1',
  httpRetryCount: 3,
  httpRetryDelay: 1000,
  limits: {
    maxCharacters: 8,
    maxMonsters: 15,
    maxTotal: 23
  },
  layout: {
    combatantPanelMaxWidth: '70%',
    combatantPanelCollapsedWidth: '16rem',
    breakpointXL: '1200px'
  },
  features: {
    gibbering: false,
    consumeLife: false
  },
  config: {
    disableIncludeLogsToggle: true,
    defaultIncludeLogs: true,
    defaultNumberOfRuns: 20,
    defaultMaxRounds: 30,
  }
};
