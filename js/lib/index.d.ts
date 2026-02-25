export interface Evaluator {
  /**
   * Get evaluator id
   */
  getId(): int;

  /**
   * Register a callback for a signal
   * @param signalName - Name of the signal to listen for
   * @param callback - Callback function to invoke when signal is emitted
   * @returns Evaluator instance for chaining
   */
  on(signalName: string, callback: (...args: any[]) => void): Evaluator;

  /**
   * Evaluate a rule with the given input
   * @param ruleName - Name of the rule to evaluate
   * @param input - Input values for the rule
   * @returns Evaluation outputs
   */
  evaluate(
    ruleName: string,
    input?: Record<string, any>,
  ): Promise<Record<string, any>>;

  /**
   * Clean up evaluator resources
   */
  destroy(): void;
}

/**
 * Create a new evaluator
 * @param programJSON - JSON string containing the compiled NIR program
 * @returns Promise resolving to an Evaluator instance
 */
export function createEvaluator(programJSON: string): Promise<Evaluator>;

/**
 * Compile a rule
 * @param src - Source code
 * @returns Promise resolving to a JSON string containing the compiled NIR program
 */
export function compile(src: string): Promise<string>;

export type Type =
  | "string"
  | "number"
  | "bool"
  | "object"
  | "any"
  | "array"
  | `array<${Type}>`
  | `${Type}|${Type}`;

export interface FunctionArgSpec {
  name: string;
  type: Type;
  required?: boolean;
}

export interface RegisterFunctionOptions {
  name: string;
  namespace?: string;
  args?: FunctionArgSpec[];
  returnType: Type;
  fn: (...args: any[]) => any;
}

/**
 * Register a custom function in the registry. Must be called before compilation or evaluation.
 * @param options - Function specification and implementation
 */
export function registerFunction(options: RegisterFunctionOptions): Promise<void>;
