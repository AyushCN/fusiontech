import * as vscode from 'vscode';
import { exec } from 'child_process';
import * as path from 'path';

export function activate(context: vscode.ExtensionContext) {
	console.log('YamlAnchor extension is now active!');

	const config = vscode.workspace.getConfiguration('yamlanchor');
	const anchorBin = config.get<string>('anchorPath', 'anchor');
	const configFile = config.get<string>('configFile', 'anchor.yaml');

	// Helper: run a CLI command in the workspace root terminal
	function runInTerminal(command: string): void {
		const terminal = vscode.window.createTerminal('YamlAnchor');
		terminal.show();
		terminal.sendText(command);
	}

	// Helper: get the current workspace root
	function getWorkspaceRoot(): string | undefined {
		return vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
	}

	// Command: Run a job locally
	const runJobDisposable = vscode.commands.registerCommand('yamlanchor.runJob', async () => {
		const root = getWorkspaceRoot();
		if (!root) {
			vscode.window.showErrorMessage('No workspace folder open.');
			return;
		}

		const jobName = await vscode.window.showInputBox({
			prompt: 'Enter the job name to run',
			placeHolder: 'e.g. build-and-test'
		});

		if (!jobName) { return; }

		runInTerminal(`cd "${root}" && ${anchorBin} local --config ${configFile}`);
		vscode.window.showInformationMessage(`⚓ Running job: ${jobName}`);
	});

	// Command: Scan for secrets
	const scanDisposable = vscode.commands.registerCommand('yamlanchor.scanSecrets', () => {
		const root = getWorkspaceRoot();
		if (!root) {
			vscode.window.showErrorMessage('No workspace folder open.');
			return;
		}
		runInTerminal(`cd "${root}" && ${anchorBin} scan .`);
		vscode.window.showInformationMessage('🔐 Scanning for secrets...');
	});

	// Command: Generate GitHub Actions YAML
	const generateDisposable = vscode.commands.registerCommand('yamlanchor.generate', () => {
		const root = getWorkspaceRoot();
		if (!root) {
			vscode.window.showErrorMessage('No workspace folder open.');
			return;
		}
		runInTerminal(`cd "${root}" && ${anchorBin} generate --config ${configFile}`);
		vscode.window.showInformationMessage('⚙️ Generating .github/workflows/main.yml...');
	});

	// Command: Open interactive exec shell
	const execDisposable = vscode.commands.registerCommand('yamlanchor.openExec', async () => {
		const root = getWorkspaceRoot();
		if (!root) {
			vscode.window.showErrorMessage('No workspace folder open.');
			return;
		}

		const jobName = await vscode.window.showInputBox({
			prompt: 'Enter the job name to exec into',
			placeHolder: 'e.g. build-and-test'
		});

		if (!jobName) { return; }

		runInTerminal(`cd "${root}" && ${anchorBin} exec ${jobName}`);
		vscode.window.showInformationMessage(`🐚 Opening shell for job: ${jobName}`);
	});

	// Code Lens Provider: Show "▶ Run" buttons next to each job in anchor.yaml
	const codeLensProvider = vscode.languages.registerCodeLensProvider(
		{ pattern: '**/anchor.yaml', language: 'yaml' },
		new AnchorCodeLensProvider(anchorBin, configFile)
	);

	context.subscriptions.push(
		runJobDisposable,
		scanDisposable,
		generateDisposable,
		execDisposable,
		codeLensProvider
	);
}

class AnchorCodeLensProvider implements vscode.CodeLensProvider {
	constructor(private anchorBin: string, private configFile: string) {}

	provideCodeLenses(document: vscode.TextDocument): vscode.CodeLens[] {
		const lenses: vscode.CodeLens[] = [];
		const text = document.getText();
		const lines = text.split('\n');

		let inJobs = false;
		for (let i = 0; i < lines.length; i++) {
			const line = lines[i];
			if (line.startsWith('jobs:')) {
				inJobs = true;
				continue;
			}
			if (inJobs && /^  \w[\w-]*:/.test(line)) {
				const jobName = line.trim().replace(':', '');
				const range = new vscode.Range(i, 0, i, line.length);

				// "▶ Run" lens
				lenses.push(new vscode.CodeLens(range, {
					title: '⚓ Run Locally',
					command: 'yamlanchor.runJob',
					arguments: [jobName]
				}));

				// "🐚 Shell" lens
				lenses.push(new vscode.CodeLens(range, {
					title: '🐚 Open Shell',
					command: 'yamlanchor.openExec',
					arguments: [jobName]
				}));
			}
		}
		return lenses;
	}
}

export function deactivate() {}
