import * as vscode from 'vscode';
import { BuddhistDebugAdapterDescriptorFactory } from './debugAdapter';

export function activate(context: vscode.ExtensionContext) {
    console.log('Buddhist Language extension is now active');

    // Register debug adapter
    const factory = new BuddhistDebugAdapterDescriptorFactory();
    context.subscriptions.push(
        vscode.debug.registerDebugAdapterDescriptorFactory('buddhist', factory)
    );
}

export function deactivate() {
    // Cleanup if needed
}
