function stripAddressPort(address) {
  const value = String(address || '').trim();
  const ipv6 = value.match(/^\[(.+)\]:\d+$/);
  if (ipv6) return ipv6[1];
  return value.replace(/:\d+$/, '');
}

export function pivotParentMap(pivotGraph) {
  const parents = new Map();

  function visit(entries, parentID = '') {
    for (const entry of entries || []) {
      const session = entry.Session || entry.session;
      const sessionID = session?.ID || session?.id;
      if (sessionID && parentID) parents.set(sessionID, parentID);
      visit(entry.Children || entry.children, sessionID || parentID);
    }
  }

  visit(pivotGraph?.Children || pivotGraph?.children);
  return parents;
}

export function agentRemoteAddress(agent, parents, agents = []) {
  const parentID = parents?.get(agent.ID);
  if (!parentID) return stripAddressPort(agent.RemoteAddress || agent.remoteAddress);

  const parent = agents instanceof Map
    ? agents.get(parentID)
    : agents.find((candidate) => candidate.ID === parentID);
  const parentAddress = stripAddressPort(parent?.RemoteAddress || parent?.remoteAddress);
  return parentAddress
    ? `${parentAddress} -> ${shortAgentID(parentID)}`
    : shortAgentID(parentID);
}

export function isAgentOnline(agent, nowSeconds = Math.floor(Date.now() / 1000)) {
  if (agent.IsDead || agent.isDead) return false;
  const kind = agent._kind || (agent.NextCheckin !== undefined ? 'beacon' : 'session');
  if (kind !== 'beacon') return true;

  const nextCheckin = Number(agent.NextCheckin ?? agent.nextCheckin ?? 0);
  if (nextCheckin <= 0) return false;

  const interval = Number(agent.Interval ?? agent.interval ?? 0) / 1e9;
  const jitter = Number(agent.Jitter ?? agent.jitter ?? 0) / 1e9;
  const grace = Math.max(15, interval + jitter);
  return nowSeconds <= nextCheckin + grace;
}

export function shortAgentID(id) {
  return String(id || '').split('-')[0];
}
