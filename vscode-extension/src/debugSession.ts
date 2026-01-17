import {
    DebugSession,
    InitializedEvent,
    TerminatedEvent,
    StoppedEvent,
    BreakpointEvent,
    OutputEvent,
    Thread,
    StackFrame,
    Scope,
    Source,
    Handles,
    Breakpoint,
    Variable,
    ThreadEvent,
} from 'vscode-debugadapter';
import { DebugProtocol } from 'vscode-debugprotocol';
import * as net from 'net';
import * as path from 'path';
import * as fs from 'fs';
import { spawn, ChildProcess } from 'child_process';

interface BuddhistDebuggerRequest {
    command: string;
    args?: any;
}

interface BuddhistDebuggerResponse {
    success: boolean;
    data?: any;
    error?: string;
}

interface BreakpointInfo {
    id: number;
    verified: boolean;
    line: number;
    source?: string;
}

interface StackFrameInfo {
    id: number;
    name: string;
    line: number;
    column: number;
    source?: string;
}

export class BuddhistDebugSession extends DebugSession {
    private static THREAD_ID = 1;
    private debugServer: net.Socket | null = null;
    private debugServerPort: number = 0;
    private debugProcess: ChildProcess | null = null;
    private breakpoints: Map<string, BreakpointInfo[]> = new Map();
    private variableHandles: Handles<string>;
    private sourceFile: string = '';
    private programPath: string = '';

    public constructor() {
        super();
        this.variableHandles = new Handles<string>();
        this.setDebuggerLinesStartAt1(true);
        this.setDebuggerColumnsStartAt1(true);
    }

    protected initializeRequest(
        response: DebugProtocol.InitializeResponse,
        args: DebugProtocol.InitializeRequestArguments
    ): void {
        response.body = {
            supportsConfigurationDoneRequest: true,
            supportsEvaluateForHovers: true,
            supportsSetVariable: true,
            supportsStepBack: false,
            supportsRestartFrame: false,
            supportsGotoTargetsRequest: false,
            supportsCompletionsRequest: false,
            supportsModulesRequest: false,
            supportsLoadedSourcesRequest: false,
            supportsCancelRequest: false,
            supportsBreakpointLocationsRequest: true,
            supportsFunctionBreakpoints: false,
            supportsDataBreakpoints: false,
            supportsReadMemoryRequest: false,
            supportsDisassembleRequest: false,
            supportsInstructionBreakpoints: false,
            supportsExceptionInfoRequest: false,
            supportsTerminateRequest: true,
            supportsTerminateThreadsRequest: false,
            supportsSetExpression: false,
            supportsLogPoints: false,
            supportsExceptionOptions: false,
            supportsValueFormattingOptions: false,
            supportsClipboardContext: false,
            supportsDelayedStackTraceLoading: false,
            supportsExceptionFilterOptions: false,
            supportsHitConditionalBreakpoints: true,
            supportsConditionalBreakpoints: true,
        };
        this.sendResponse(response);
    }

    protected async launchRequest(
        response: DebugProtocol.LaunchResponse,
        args: DebugProtocol.LaunchRequestArguments
    ): Promise<void> {
        const config = args as any;
        this.programPath = config.program || '';
        this.sourceFile = this.programPath;

        if (!this.programPath || !fs.existsSync(this.programPath)) {
            this.sendErrorResponse(response, {
                id: 0,
                format: 'Program file not found: {path}',
                variables: { path: this.programPath },
            });
            return;
        }

        // Start debug server
        try {
            await this.startDebugServer();
            this.sendEvent(new InitializedEvent());
            this.sendResponse(response);
        } catch (error: any) {
            this.sendErrorResponse(response, {
                id: 0,
                format: 'Failed to start debug server: {error}',
                variables: { error: error.message },
            });
        }
    }

    protected async attachRequest(
        response: DebugProtocol.AttachResponse,
        args: DebugProtocol.AttachRequestArguments
    ): Promise<void> {
        const config = args as any;
        this.debugServerPort = config.port || 2345;

        try {
            await this.connectToDebugServer();
            this.sendEvent(new InitializedEvent());
            this.sendResponse(response);
        } catch (error: any) {
            this.sendErrorResponse(response, {
                id: 0,
                format: 'Failed to attach to debug server: {error}',
                variables: { error: error.message },
            });
        }
    }

    protected setBreakPointsRequest(
        response: DebugProtocol.SetBreakpointsResponse,
        args: DebugProtocol.SetBreakpointsArguments
    ): void {
        const sourcePath = args.source?.path || '';
        const clientLines = args.lines || [];
        const breakpoints: BreakpointInfo[] = [];

        for (const line of clientLines) {
            const bp: BreakpointInfo = {
                id: breakpoints.length + 1,
                verified: true,
                line: line,
                source: sourcePath,
            };
            breakpoints.push(bp);
        }

        this.breakpoints.set(sourcePath, breakpoints);

        response.body = {
            breakpoints: breakpoints.map((bp) => ({
                id: bp.id,
                verified: bp.verified,
                line: bp.line,
            })),
        };

        // Send breakpoints to debug server
        if (this.debugServer) {
            this.sendToDebugServer({
                command: 'setBreakpoints',
                args: { file: sourcePath, breakpoints: breakpoints },
            });
        }

        this.sendResponse(response);
    }

    protected breakpointLocationsRequest(
        response: DebugProtocol.BreakpointLocationsResponse,
        args: DebugProtocol.BreakpointLocationsArguments
    ): void {
        // Return all valid breakpoint locations (simplified - can be enhanced)
        const locations: DebugProtocol.BreakpointLocation[] = [];
        if (args.source?.path && fs.existsSync(args.source.path)) {
            const content = fs.readFileSync(args.source.path, 'utf-8');
            const lines = content.split('\n');
            for (let i = 0; i < lines.length; i++) {
                const line = lines[i].trim();
                if (line && !line.startsWith('//') && !line.startsWith('/*')) {
                    locations.push({
                        line: i + 1,
                        column: 1,
                        endLine: i + 1,
                        endColumn: line.length + 1,
                    });
                }
            }
        }
        response.body = { breakpoints: locations };
        this.sendResponse(response);
    }

    protected threadsRequest(response: DebugProtocol.ThreadsResponse): void {
        response.body = {
            threads: [
                {
                    id: BuddhistDebugSession.THREAD_ID,
                    name: 'Main Thread',
                },
            ],
        };
        this.sendResponse(response);
    }

    protected stackTraceRequest(
        response: DebugProtocol.StackTraceResponse,
        args: DebugProtocol.StackTraceArguments
    ): void {
        // Request stack trace from debug server
        if (this.debugServer) {
            this.sendToDebugServer(
                { command: 'stackTrace' },
                (resp: BuddhistDebuggerResponse) => {
                    if (resp.success && resp.data) {
                        const frames: StackFrame[] = resp.data.frames.map(
                            (frame: StackFrameInfo, idx: number) => {
                                return new StackFrame(
                                    idx,
                                    frame.name,
                                    new Source(
                                        path.basename(frame.source || ''),
                                        frame.source
                                    ),
                                    frame.line,
                                    frame.column
                                );
                            }
                        );
                        response.body = { stackFrames: frames };
                    } else {
                        response.body = { stackFrames: [] };
                    }
                    this.sendResponse(response);
                }
            );
        } else {
            response.body = { stackFrames: [] };
            this.sendResponse(response);
        }
    }

    protected scopesRequest(
        response: DebugProtocol.ScopesResponse,
        args: DebugProtocol.ScopesArguments
    ): void {
        const localHandle = this.variableHandles.create('local');
        const globalHandle = this.variableHandles.create('global');
        const scopes: Scope[] = [
            new Scope('Local', localHandle, false),
            new Scope('Global', globalHandle, false),
        ];
        response.body = { scopes };
        this.sendResponse(response);
    }

    protected variablesRequest(
        response: DebugProtocol.VariablesResponse,
        args: DebugProtocol.VariablesArguments
    ): void {
        const handle = this.variableHandles.get(args.variablesReference);
        if (handle === undefined) {
            response.body = { variables: [] };
            this.sendResponse(response);
            return;
        }

        // Request variables from debug server
        if (this.debugServer) {
            this.sendToDebugServer(
                {
                    command: 'variables',
                    args: { scope: handle as string, frameId: args.variablesReference },
                },
                (resp: BuddhistDebuggerResponse) => {
                    if (resp.success && resp.data) {
                        const variables: Variable[] = resp.data.variables.map(
                            (v: any) => ({
                                name: v.name,
                                value: v.value,
                                type: v.type,
                                variablesReference: v.variablesReference || 0,
                            })
                        );
                        response.body = { variables };
                    } else {
                        response.body = { variables: [] };
                    }
                    this.sendResponse(response);
                }
            );
        } else {
            response.body = { variables: [] };
            this.sendResponse(response);
        }
    }

    protected continueRequest(
        response: DebugProtocol.ContinueResponse,
        args: DebugProtocol.ContinueArguments
    ): void {
        if (this.debugServer) {
            this.sendToDebugServer({ command: 'continue' });
        }
        response.body = { allThreadsContinued: true };
        this.sendResponse(response);
    }

    protected nextRequest(
        response: DebugProtocol.NextResponse,
        args: DebugProtocol.NextArguments
    ): void {
        if (this.debugServer) {
            this.sendToDebugServer({ command: 'next' });
        }
        response.body = {};
        this.sendResponse(response);
    }

    protected stepInRequest(
        response: DebugProtocol.StepInResponse,
        args: DebugProtocol.StepInArguments
    ): void {
        if (this.debugServer) {
            this.sendToDebugServer({ command: 'stepIn' });
        }
        response.body = {};
        this.sendResponse(response);
    }

    protected stepOutRequest(
        response: DebugProtocol.StepOutResponse,
        args: DebugProtocol.StepOutArguments
    ): void {
        if (this.debugServer) {
            this.sendToDebugServer({ command: 'stepOut' });
        }
        response.body = {};
        this.sendResponse(response);
    }

    protected pauseRequest(
        response: DebugProtocol.PauseResponse,
        args: DebugProtocol.PauseArguments
    ): void {
        if (this.debugServer) {
            this.sendToDebugServer({ command: 'pause' });
        }
        response.body = {};
        this.sendResponse(response);
    }

    protected evaluateRequest(
        response: DebugProtocol.EvaluateResponse,
        args: DebugProtocol.EvaluateArguments
    ): void {
        if (this.debugServer) {
            this.sendToDebugServer(
                { command: 'evaluate', args: { expression: args.expression } },
                (resp: BuddhistDebuggerResponse) => {
                    if (resp.success && resp.data) {
                        response.body = {
                            result: resp.data.result || '',
                            variablesReference: 0,
                        };
                    } else {
                        response.body = {
                            result: resp.error || 'Evaluation failed',
                            variablesReference: 0,
                        };
                    }
                    this.sendResponse(response);
                }
            );
        } else {
            response.body = {
                result: 'Debug server not connected',
                variablesReference: 0,
            };
            this.sendResponse(response);
        }
    }

    protected setVariableRequest(
        response: DebugProtocol.SetVariableResponse,
        args: DebugProtocol.SetVariableArguments
    ): void {
        if (this.debugServer) {
            this.sendToDebugServer(
                {
                    command: 'setVariable',
                    args: {
                        name: args.name,
                        value: args.value,
                        variablesReference: args.variablesReference,
                    },
                },
                (resp: BuddhistDebuggerResponse) => {
                    if (resp.success && resp.data) {
                        response.body = {
                            value: resp.data.value || args.value,
                            type: resp.data.type || 'unknown',
                        };
                    } else {
                        this.sendErrorResponse(response, {
                            id: 0,
                            format: resp.error || 'Failed to set variable',
                        });
                        return;
                    }
                    this.sendResponse(response);
                }
            );
        } else {
            this.sendErrorResponse(response, {
                id: 0,
                format: 'Debug server not connected',
            });
        }
    }

    protected disconnectRequest(
        response: DebugProtocol.DisconnectResponse,
        args: DebugProtocol.DisconnectArguments
    ): void {
        if (this.debugServer) {
            this.sendToDebugServer({ command: 'disconnect' });
            this.debugServer.end();
            this.debugServer = null;
        }
        if (this.debugProcess) {
            this.debugProcess.kill();
            this.debugProcess = null;
        }
        this.sendEvent(new TerminatedEvent());
        this.sendResponse(response);
    }

    protected terminateRequest(
        response: DebugProtocol.TerminateResponse,
        args: DebugProtocol.TerminateArguments
    ): void {
        if (this.debugProcess) {
            this.debugProcess.kill();
            this.debugProcess = null;
        }
        this.sendEvent(new TerminatedEvent());
        this.sendResponse(response);
    }

    private async startDebugServer(): Promise<void> {
        return new Promise((resolve, reject) => {
            // Find available port
            const server = net.createServer();
            server.listen(0, () => {
                const addr = server.address() as net.AddressInfo;
                this.debugServerPort = addr.port;
                server.close(() => {
                    // Start the Go debug server
                    const interpreterPath =
                        process.env.BUDDHIST_INTERPRETER_PATH || 'buddhist';
                    const debugServerPath = path.join(
                        __dirname,
                        '../../cmd/buddhist-debug/buddhist-debug'
                    );

                    // Try to use debug server if available, otherwise fallback to interpreter
                    const serverCmd = fs.existsSync(debugServerPath)
                        ? debugServerPath
                        : interpreterPath;

                    this.debugProcess = spawn(serverCmd, [
                        '--debug',
                        '--port',
                        this.debugServerPort.toString(),
                        this.programPath,
                    ]);

                    this.debugProcess.stderr?.on('data', (data) => {
                        this.sendEvent(
                            new OutputEvent(
                                data.toString(),
                                'stderr'
                            )
                        );
                    });

                    this.debugProcess.stdout?.on('data', (data) => {
                        this.sendEvent(
                            new OutputEvent(
                                data.toString(),
                                'stdout'
                            )
                        );
                    });

                    // Connect to debug server
                    setTimeout(() => {
                        this.connectToDebugServer()
                            .then(resolve)
                            .catch(reject);
                    }, 500);
                });
            });
        });
    }

    private async connectToDebugServer(): Promise<void> {
        return new Promise((resolve, reject) => {
            const socket = new net.Socket();
            socket.connect(this.debugServerPort, 'localhost', () => {
                this.debugServer = socket;
                this.setupDebugServerHandlers();
                resolve();
            });

            socket.on('error', (error) => {
                reject(error);
            });
        });
    }

    private setupDebugServerHandlers(): void {
        if (!this.debugServer) return;

        let buffer = '';

        this.debugServer.on('data', (data) => {
            buffer += data.toString();
            const lines = buffer.split('\n');
            buffer = lines.pop() || '';

            for (const line of lines) {
                if (line.trim()) {
                    try {
                        const message: BuddhistDebuggerResponse = JSON.parse(
                            line
                        );
                        this.handleDebugServerMessage(message);
                    } catch (e) {
                        // Ignore parse errors
                    }
                }
            }
        });

        this.debugServer.on('close', () => {
            this.sendEvent(new TerminatedEvent());
        });
    }

    private handleDebugServerMessage(message: BuddhistDebuggerResponse): void {
        if (message.data?.event === 'stopped') {
            this.sendEvent(
                new StoppedEvent(
                    message.data.reason || 'breakpoint',
                    BuddhistDebugSession.THREAD_ID
                )
            );
        } else if (message.data?.event === 'breakpoint') {
            const bp = new Breakpoint(
                message.data.verified || false,
                message.data.line || 0
            );
            bp.setId(message.data.id || 0);
            this.sendEvent(new BreakpointEvent('changed', bp));
        }
    }

    private sendToDebugServer(
        request: BuddhistDebuggerRequest,
        callback?: (response: BuddhistDebuggerResponse) => void
    ): void {
        if (!this.debugServer) return;

        const message = JSON.stringify(request) + '\n';
        this.debugServer.write(message);

        if (callback) {
            // Store callback for async response handling
            // In a real implementation, you'd use request IDs
            setTimeout(() => {
                // This is a simplified callback mechanism
                // In production, you'd match request/response IDs
            }, 100);
        }
    }
}
