import type { IntegrationCatalogItem, PostInstallPlan } from './integrations';

export const SDK_LANGUAGES = ['node', 'python', 'go'] as const;
export type SdkLanguage = (typeof SDK_LANGUAGES)[number];
export const SHELLS = ['posix', 'powershell'] as const;
export type Shell = (typeof SHELLS)[number];
type SecretDelivery = 'embedded' | 'environment';

interface AgentInstallIdentity {
  readonly agentId: string;
  readonly kid: string;
  readonly orgId: string;
}

interface InstructionInput {
  readonly integration: IntegrationCatalogItem;
  readonly identity: AgentInstallIdentity;
  readonly baseUrl: string;
  readonly privateKey: string;
  readonly token: string;
}

export interface InstallInstructions {
  readonly setup: string;
  readonly usage?: string;
  readonly verify?: string;
  readonly postInstall?: PostInstallPlan;
  readonly secretDelivery: SecretDelivery;
}

function shellQuoted(value: string, label: string): string {
  if (!value || /['\u0000-\u001f\u007f]/u.test(value)) {
    throw new Error(`${label} cannot be represented in the generated shell command.`);
  }
  return `'${value}'`;
}

const INSTALLER: Record<SdkLanguage, string> = {
  node: 'npx @elydora/sdk install',
  python: 'elydora install',
  go: 'elydora install',
};
const BOOTSTRAP: Partial<Record<SdkLanguage, string>> = {
  python: 'python -m pip install elydora',
  go: 'go install github.com/Elydora-Infra/Elydora-Go-SDK/v2/cmd/elydora@latest',
};

interface CommandValues {
  readonly agent: string;
  readonly orgId: string;
  readonly agentId: string;
  readonly kid: string;
  readonly baseUrl: string;
}

function credentialFileStem(agentId: string): string {
  if (!/^[A-Za-z0-9._-]+$/u.test(agentId)) {
    throw new Error('Agent ID cannot be used as a credential file name.');
  }
  return `.elydora-${agentId}`;
}

function installArguments(
  language: SdkLanguage,
  values: CommandValues,
  keyFile: string,
  tokenFile: string,
): string {
  if (language === 'go') {
    return `--agent ${values.agent} --org-id ${values.orgId} --agent-id ${values.agentId} --kid ${values.kid} --base-url ${values.baseUrl} --private-key-file ${keyFile} --token-file ${tokenFile}`;
  }
  return `--agent ${values.agent} --org_id ${values.orgId} --agent_id ${values.agentId} --kid ${values.kid} --base_url ${values.baseUrl} --private_key_file ${keyFile} --token_file ${tokenFile}`;
}

function posixInstall(
  language: SdkLanguage,
  values: CommandValues,
  stem: string,
  privateKey: string,
  token: string,
): string {
  const template = `"$HOME/${stem}.XXXXXX"`;
  const script = [
    `k=$(mktemp ${template})`,
    `trap 'rm -f "$k"' EXIT INT TERM`,
    `t=$(mktemp ${template})`,
    `trap 'rm -f "$k" "$t"' EXIT INT TERM`,
    `printf '%s\\n' ${shellQuoted(privateKey, 'Private key')} > "$k"`,
    `printf '%s\\n' ${shellQuoted(token, 'API token')} > "$t"`,
    `${INSTALLER[language]} ${installArguments(language, values, '"$k"', '"$t"')}`,
  ].join(' && ');
  return `printf '%s' '${script.replace(/'/gu, "'\\''")}' | sh`;
}

function powershellInstall(
  language: SdkLanguage,
  values: CommandValues,
  stem: string,
  privateKey: string,
  token: string,
): string {
  const paths = `$k = Join-Path $HOME '${stem}.key'; $t = Join-Path $HOME '${stem}.token'`;
  const create = 'New-Item -ItemType File -Force -Path $k, $t | Out-Null';
  const protect = 'if ($IsLinux -or $IsMacOS) { chmod 600 $k $t; if ($LASTEXITCODE -ne 0) { throw "chmod failed" } } else { $u = [System.Security.Principal.WindowsIdentity]::GetCurrent().User.Value; foreach ($f in $k, $t) { icacls $f /inheritance:r /grant:r "*$($u):F" | Out-Null; if ($LASTEXITCODE -ne 0) { throw "icacls failed for $f" } } }';
  const store = `Set-Content -LiteralPath $k -Value ${shellQuoted(privateKey, 'Private key')}; Set-Content -LiteralPath $t -Value ${shellQuoted(token, 'API token')}`;
  const install = `${INSTALLER[language]} ${installArguments(language, values, '$k', '$t')}`;
  const status = 'if ($LASTEXITCODE -ne 0) { throw "elydora install exited with code $LASTEXITCODE" }';
  return `${paths}; try { ${create}; ${protect}; ${store}; ${install}; ${status} } finally { Remove-Item -LiteralPath $k, $t -Force -ErrorAction SilentlyContinue }`;
}

function hookInstallerCommand(
  language: SdkLanguage,
  { integration, identity, baseUrl, privateKey, token }: InstructionInput,
  shell: Shell,
): string {
  const values: CommandValues = {
    agent: integration.id,
    orgId: shellQuoted(identity.orgId, 'Organization ID'),
    agentId: shellQuoted(identity.agentId, 'Agent ID'),
    kid: shellQuoted(identity.kid, 'Key ID'),
    baseUrl: shellQuoted(baseUrl, 'API base URL'),
  };
  const stem = credentialFileStem(identity.agentId);
  const command = shell === 'powershell'
    ? powershellInstall(language, values, stem, privateKey, token)
    : posixInstall(language, values, stem, privateKey, token);
  const bootstrap = BOOTSTRAP[language];
  return bootstrap ? `${bootstrap}\n${command}` : command;
}

function sdkSetup(language: SdkLanguage, input: InstructionInput): string {
  const { identity, baseUrl } = input;
  if (language === 'python') {
    return `python -m pip install elydora

import os

from elydora import ElydoraClient

def require_env(name: str) -> str:
    value = os.environ.get(name)
    if not value:
        raise RuntimeError(f"{name} is required")
    return value

client = ElydoraClient(
    org_id=${JSON.stringify(identity.orgId)},
    agent_id=${JSON.stringify(identity.agentId)},
    private_key=require_env("ELYDORA_PRIVATE_KEY"),
    base_url=${JSON.stringify(baseUrl)},
    token=require_env("ELYDORA_API_TOKEN"),
)
client.set_kid(${JSON.stringify(identity.kid)})`;
  }

  if (language === 'go') {
    return `go get github.com/Elydora-Infra/Elydora-Go-SDK/v2

requireEnv := func(name string) string {
    value, ok := os.LookupEnv(name)
    if !ok || value == "" {
        panic(name + " is required")
    }
    return value
}

client, err := elydora.NewClient(&elydora.Config{
    OrgID: ${JSON.stringify(identity.orgId)},
    AgentID: ${JSON.stringify(identity.agentId)},
    PrivateKey: requireEnv("ELYDORA_PRIVATE_KEY"),
    BaseURL: ${JSON.stringify(baseUrl)},
    Token: requireEnv("ELYDORA_API_TOKEN"),
})
if err != nil {
    panic(err)
}`;
  }

  return `npm install @elydora/sdk

import { ElydoraClient } from '@elydora/sdk';

const privateKey = process.env.ELYDORA_PRIVATE_KEY;
const token = process.env.ELYDORA_API_TOKEN;
if (!privateKey || !token) {
  throw new Error('ELYDORA_PRIVATE_KEY and ELYDORA_API_TOKEN are required');
}

const client = new ElydoraClient({
  orgId: ${JSON.stringify(identity.orgId)},
  agentId: ${JSON.stringify(identity.agentId)},
  privateKey,
  kid: ${JSON.stringify(identity.kid)},
  baseUrl: ${JSON.stringify(baseUrl)},
});
client.setToken(token);`;
}

function sdkUsage(language: SdkLanguage): string {
  if (language === 'python') {
    return `operation = client.create_operation(
    operation_type="ai.tool_use",
    subject={"session_id": "session-123"},
    action={"tool": "shell"},
    payload={"command": "example"},
)
receipt = client.submit_operation(operation)`;
  }

  if (language === 'go') {
    return `operation, err := client.CreateOperation(&elydora.CreateOperationParams{
    OperationType: "ai.tool_use",
    Subject: map[string]any{"session_id": "session-123"},
    Action: map[string]any{"tool": "shell"},
    Payload: map[string]any{"command": "example"},
})
if err != nil {
    panic(err)
}
_, err = client.SubmitOperation(operation)
if err != nil {
    panic(err)
}`;
  }

  return `const operation = client.createOperation({
  operationType: 'ai.tool_use',
  subject: { session_id: 'session-123' },
  action: { tool: 'shell' },
  payload: { command: 'example' },
});
await client.submitOperation(operation);`;
}

export function buildInstallInstructions(
  language: SdkLanguage,
  input: InstructionInput,
  shell: Shell = 'posix',
): InstallInstructions {
  if (input.integration.mode === 'hooks') {
    return {
      setup: hookInstallerCommand(language, input, shell),
      verify: language === 'node' ? 'npx @elydora/sdk status' : 'elydora status',
      postInstall: input.integration.postInstall,
      secretDelivery: 'embedded',
    };
  }

  return {
    setup: sdkSetup(language, input),
    usage: sdkUsage(language),
    secretDelivery: 'environment',
  };
}
