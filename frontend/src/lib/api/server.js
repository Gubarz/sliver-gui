import {
  GetCredentials,
  GetImplantBuilds,
  GetJobs,
  GetLoot,
  GetOperators,
  GetProfiles,
} from '../../../wailsjs/go/main/App.js';
import { responseField } from './normalize.js';

export {
  AddCredential,
  DeleteImplantBuild,
  DeleteProfile,
  DownloadLoot,
  GenerateImplant,
  GenerateImplantFromProfile,
  GetPivots,
  GetPivotListeners,
  GetScreenshotData,
  KillJob,
  RegenerateImplant,
  RemoveCredential,
  RemoveLoot,
  SaveProfile,
  StartListener,
} from '../../../wailsjs/go/main/App.js';

export async function listCredentials() {
  return responseField(await GetCredentials(), 'Credentials', []);
}

export async function listJobs() {
  return responseField(await GetJobs(), 'Active', []);
}

export async function listLoot() {
  return responseField(await GetLoot(), 'Loot', []);
}

export async function listOperators() {
  return responseField(await GetOperators(), 'Operators', []);
}

export async function listProfiles() {
  return responseField(await GetProfiles(), 'Profiles', []);
}

export async function listImplantBuilds() {
  const configs = responseField(await GetImplantBuilds(), 'Configs', {});
  return Object.entries(configs).map(([name, config]) => ({ name, ...config }));
}
