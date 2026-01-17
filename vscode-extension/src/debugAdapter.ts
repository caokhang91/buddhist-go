import * as vscode from 'vscode';
import * as path from 'path';
import { DebugAdapterExecutable, DebugAdapterServer, DebugAdapterDescriptor, ProviderResult } from 'vscode';
import { BuddhistDebugSession } from './debugSession';

export class BuddhistDebugAdapterDescriptorFactory implements vscode.DebugAdapterDescriptorFactory {
    createDebugAdapterDescriptor(
        session: vscode.DebugSession,
        executable: DebugAdapterExecutable | undefined
    ): ProviderResult<DebugAdapterDescriptor> {
        // Use in-process debug adapter
        return new vscode.DebugAdapterInlineImplementation(new BuddhistDebugSession());
    }
}
