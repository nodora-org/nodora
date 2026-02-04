%{
package parser

import (
	"fmt"
    "nodora.org/nodora/internal/ast"
)

var parseResult *ast.Program

%}

%union {
    str     string
    node     ast.Node
    nodes    []ast.Node
    signal  *ast.Signal
    rule    *ast.Rule
    param   ast.Param
    params  []ast.Param
    stmt    ast.Statement
    stmts   []ast.Statement
    expr    ast.Expr
    exprs   []ast.Expr
}

%token <str> IDENT STRING NUMBER
%token TRUE FALSE
%token SIGNAL RULE EMIT WHEN OUT
%token LPAREN RPAREN LBRACE RBRACE LBRACKET RBRACKET COLON COMMA QMARK DOT
%token PLUS MINUS STAR SLASH MOD
%token GT LT GTE LTE EQ NEQ AND OR NOT ASSIGN IN
%token IF THEN ELSE

%type <node> decl
%type <nodes> decl_list
%type <signal> signal_decl
%type <rule> rule_decl
%type <param> param
%type <params> param_list params_opt
%type <stmt> stmt assign_stmt emit_stmt
%type <stmts> stmt_list
%type <expr> expr conditional_expr
%type <expr> logical_or_expr logical_and_expr equality_expr relational_expr membership_expr additive_expr multiplicative_expr unary_expr postfix_expr primary_expr
%type <exprs> arg_list args_opt

%right QMARK
%left OR
%left AND
%left EQ NEQ
%left IN
%left GT LT GTE LTE
%left PLUS MINUS
%left STAR SLASH MOD
%right NOT

%%

program
    : decl_list
        { parseResult = &ast.Program{Decls: $1} }
    ;

decl_list
    : decl_list decl
        { $$ = append($1, $2) }
    | decl
        { $$ = []ast.Node{$1} }
    ;

decl
    : signal_decl
        { $$ = $1 }
    | rule_decl
        { $$ = $1 }
    ;

signal_decl
    : SIGNAL IDENT LPAREN params_opt RPAREN
        { $$ = &ast.Signal{Name: $2, Params: $4} }
    ;

rule_decl
    : RULE IDENT LBRACE stmt_list RBRACE
        { $$ = &ast.Rule{Name: $2, Statements: $4} }
    ;

param
    : IDENT
        { $$ = ast.Param{Name: $1} }
    ;

param_list
    : param
        { $$ = []ast.Param{$1} }
    | param_list COMMA param
        { $$ = append($1, $3) }
    ;

params_opt
    :
        { $$ = []ast.Param{} }
    | param_list
        { $$ = $1 }
    ;

stmt_list
    : /* empty */
        { $$ = []ast.Statement{} }
    | stmt_list stmt
        { $$ = append($1, $2) }
    ;

stmt
    : assign_stmt
        { $$ = $1 }
    | emit_stmt
        { $$ = $1 }
    ;

assign_stmt
    : IDENT ASSIGN expr
        { $$ = &ast.Assignment{Name: $1, Expr: $3, IsOut: false} }
    | OUT IDENT ASSIGN expr
        { $$ = &ast.Assignment{Name: $2, Expr: $4, IsOut: true} }
    ;

emit_stmt
    : EMIT IDENT LPAREN args_opt RPAREN WHEN expr
        { $$ = &ast.EmitStatement{Signal: $2, Args: $4, Condition: $7} }
    | EMIT IDENT LPAREN args_opt RPAREN
        { $$ = &ast.EmitStatement{Signal: $2, Args: $4} }
    ;

expr
    : conditional_expr
        { $$ = $1 }
    ;

conditional_expr
    : logical_or_expr
    | IF expr THEN expr ELSE conditional_expr
        { $$ = &ast.ConditionalExpr{Cond: $2, Then: $4, Else: $6} }
    | logical_or_expr QMARK expr COLON conditional_expr
        { $$ = &ast.ConditionalExpr{Cond: $1, Then: $3, Else: $5} }
    ;

logical_or_expr
    : logical_or_expr OR logical_and_expr
        { $$ = &ast.BinaryExpr{Left: $1, Op: "||", Right: $3} }
    | logical_and_expr
    ;

logical_and_expr
    : logical_and_expr AND equality_expr
        { $$ = &ast.BinaryExpr{Left: $1, Op: "&&", Right: $3} }
    | equality_expr
    ;

equality_expr
    : equality_expr EQ relational_expr
        { $$ = &ast.BinaryExpr{Left: $1, Op: "==", Right: $3} }
    | equality_expr NEQ relational_expr
        { $$ = &ast.BinaryExpr{Left: $1, Op: "!=", Right: $3} }
    | relational_expr
    ;

relational_expr
    : relational_expr GT membership_expr
        { $$ = &ast.BinaryExpr{Left: $1, Op: ">", Right: $3} }
    | relational_expr LT membership_expr
        { $$ = &ast.BinaryExpr{Left: $1, Op: "<", Right: $3} }
    | relational_expr GTE membership_expr
        { $$ = &ast.BinaryExpr{Left: $1, Op: ">=", Right: $3} }
    | relational_expr LTE membership_expr
        { $$ = &ast.BinaryExpr{Left: $1, Op: "<=", Right: $3} }
    | membership_expr
    ;

membership_expr
    : membership_expr IN postfix_expr
        { $$ = &ast.BinaryExpr{Left: $1, Op: "in", Right: $3} }
    | additive_expr
    ;

additive_expr
    : additive_expr PLUS multiplicative_expr
        { $$ = &ast.BinaryExpr{Left: $1, Op: "+", Right: $3} }
    | additive_expr MINUS multiplicative_expr
        { $$ = &ast.BinaryExpr{Left: $1, Op: "-", Right: $3} }
    | multiplicative_expr MOD unary_expr
        { $$ = &ast.BinaryExpr{Left: $1, Op: "%", Right: $3} }
    | multiplicative_expr
    ;

multiplicative_expr
    : multiplicative_expr STAR unary_expr
        { $$ = &ast.BinaryExpr{Left: $1, Op: "*", Right: $3} }
    | multiplicative_expr SLASH unary_expr
        { $$ = &ast.BinaryExpr{Left: $1, Op: "/", Right: $3} }
    | unary_expr
    ;

unary_expr
    : postfix_expr
        { $$ = $1 }
    | NOT unary_expr
        { $$ = &ast.UnaryExpr{Op: "!", Expr: $2} }
    | MINUS unary_expr
        { $$ = &ast.UnaryExpr{Op: "-", Expr: $2} }
    ;

postfix_expr
    : primary_expr
        { $$ = $1 }
    | postfix_expr DOT IDENT
        { $$ = &ast.SelectorExpr{Expr: $1, Field: $3} }
    ;

primary_expr
    : IDENT
        { $$ = &ast.Identifier{Name: $1} }
    | NUMBER
        { $$ = &ast.NumberLiteral{Value: $1} }
    | STRING
        { $$ = &ast.StringLiteral{Value: $1} }
    | TRUE
        { $$ = &ast.BoolLiteral{Value: true} }
    | FALSE
        { $$ = &ast.BoolLiteral{Value: false} }
    | LBRACKET args_opt RBRACKET
        { $$ = &ast.ArrayLiteral{Elements: $2} }
    | LPAREN expr RPAREN
        { $$ = $2 }
    ;

args_opt
    :
        { $$ = []ast.Expr{} }
    | arg_list
        { $$ = $1 }
    ;

arg_list
    : expr
        { $$ = []ast.Expr{$1} }
    | arg_list COMMA expr
        { $$ = append($1, $3) }
    ;

%%

func Parse(input string) (*ast.Program, error) {
	l := NewLexer(input)
	p := yyNewParser()
	p.Parse(l)
	if l.lastError != "" {
		return nil, fmt.Errorf("%d:%d : %s", l.line, l.col, l.lastError)
	}
	return parseResult, nil
}