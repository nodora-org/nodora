%{
package parser

import (
	"fmt"
    "nodora.org/nodora/internal/ast"
)

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
    span    ast.Span
    prop    ast.ObjectProperty
    props   []ast.ObjectProperty
    arm     *ast.MatchArm
    arms    []*ast.MatchArm
}

%token <str> IDENT STRING NUMBER
%token <span> TRUE FALSE SIGNAL RULE EMIT WHEN OUT IN IF THEN ELSE MATCH
%token <span> LPAREN RPAREN LBRACE RBRACE LBRACKET RBRACKET COLON COMMA QMARK DOT NAMESPACE
%token <span> PLUS MINUS STAR SLASH MOD
%token <span> GT LT GTE LTE EQ NEQ AND OR NOT ASSIGN FATARROW
%token <span> PIPE

%type <node> decl
%type <nodes> decl_list
%type <signal> signal_decl
%type <rule> rule_decl
%type <param> param
%type <params> param_list params_opt
%type <stmt> stmt assign_stmt emit_stmt
%type <stmts> stmt_list
%type <expr> expr conditional_expr match_expr match_pattern
%type <expr> logical_or_expr logical_and_expr equality_expr relational_expr membership_expr additive_expr multiplicative_expr unary_expr postfix_expr primary_expr
%type <exprs> arg_list args_opt
%type <prop> obj_prop
%type <props> obj_prop_list
%type <arm> match_arm
%type <arms> match_arms

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
        { yylex.(*lexer).result = &ast.Program{Decls: $1} }
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
        { 
          $$ = &ast.Signal{Name: $2, Params: $4}
          $$.Span = $1.Merge($5)
        }
    ;

rule_decl
    : RULE IDENT LBRACE stmt_list RBRACE
        { 
          $$ = &ast.Rule{Name: $2, Statements: $4}
          $$.Span = $1.Merge($5)
        }
    ;

param
    : IDENT
        { 
          $$ = ast.Param{Name: $1}
          $$.Span = $<span>1
        }
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
        { 
          a := &ast.Assignment{Name: $1, Expr: $3, IsOut: false}
          a.Span = $<span>1.Merge($3.GetSpan())
          $$ = a
        }
    | OUT IDENT ASSIGN expr
        { 
          a := &ast.Assignment{Name: $2, Expr: $4, IsOut: true}
          a.Span = $1.Merge($4.GetSpan())
          $$ = a
        }
    ;

emit_stmt
    : EMIT IDENT LPAREN args_opt RPAREN WHEN expr
        { 
          e := &ast.EmitStatement{Signal: $2, Args: $4, Condition: $7}
          e.Span = $1.Merge($7.GetSpan())
          $$ = e
        }
    | EMIT IDENT LPAREN args_opt RPAREN
        { 
          e := &ast.EmitStatement{Signal: $2, Args: $4}
          e.Span = $1.Merge($5)
          $$ = e
        }
    ;

expr
    : PIPE param_list PIPE expr
        { 
          le := &ast.LambdaExpr{Params: $2, Body: $4}
          le.Span = $1.Merge($4.GetSpan())
          $$ = le
        }
    | conditional_expr
        { $$ = $1 }
    ;

conditional_expr
    : logical_or_expr
        { $$ = $1 }
    | IF expr THEN expr ELSE conditional_expr
        {
          c := &ast.ConditionalExpr{Cond: $2, Then: $4, Else: $6}
          c.Span = $1.Merge($6.GetSpan())
          $$ = c
        }
    | logical_or_expr QMARK expr COLON conditional_expr
        {
          c := &ast.ConditionalExpr{Cond: $1, Then: $3, Else: $5}
          c.Span = $1.GetSpan().Merge($5.GetSpan())
          $$ = c
        }
    | match_expr
        { $$ = $1 }
    ;

match_expr
    : MATCH logical_or_expr LBRACE match_arms RBRACE
        {
          m := &ast.MatchExpr{Value: $2, Arms: $4}
          m.Span = $1.Merge($5)
          $$ = m
        }
    | MATCH logical_or_expr LBRACE match_arms COMMA RBRACE
        {
          m := &ast.MatchExpr{Value: $2, Arms: $4}
          m.Span = $1.Merge($6)
          $$ = m
        }
    ;

match_arms
    : match_arm
        { $$ = []*ast.MatchArm{$1} }
    | match_arms COMMA match_arm
        { $$ = append($1, $3) }
    ;

match_arm
    : match_pattern FATARROW expr
        {
          a := &ast.MatchArm{Pattern: $1, Body: $3}
          a.Span = $1.GetSpan().Merge($3.GetSpan())
          $$ = a
        }
    | match_pattern WHEN expr FATARROW expr
        {
          a := &ast.MatchArm{Pattern: $1, Guard: $3, Body: $5}
          a.Span = $1.GetSpan().Merge($5.GetSpan())
          $$ = a
        }
    ;

match_pattern
    : IDENT
        {
          i := &ast.Identifier{Name: $1}
          i.Span = $<span>1
          $$ = i
        }
    | NUMBER
        {
          n := &ast.NumberLiteral{Raw: $1}
          n.Span = $<span>1
          $$ = n
        }
    | MINUS NUMBER
        {
          n := &ast.NumberLiteral{Raw: $2}
          n.Span = $<span>2
          u := &ast.UnaryExpr{Op: "-", Expr: n}
          u.Span = $1.Merge($<span>2)
          $$ = u
        }
    | STRING
        {
          s := &ast.StringLiteral{Value: $1}
          s.Span = $<span>1
          $$ = s
        }
    | TRUE
        {
          b := &ast.BoolLiteral{Value: true}
          b.Span = $1
          $$ = b
        }
    | FALSE
        {
          b := &ast.BoolLiteral{Value: false}
          b.Span = $1
          $$ = b
        }
    ;

logical_or_expr
    : logical_or_expr OR logical_and_expr
        { 
          b := &ast.BinaryExpr{Left: $1, Op: "||", Right: $3}
          b.Span = $1.GetSpan().Merge($3.GetSpan())
          $$ = b
        }
    | logical_and_expr
        { $$ = $1 }
    ;

logical_and_expr
    : logical_and_expr AND equality_expr
        { 
          b := &ast.BinaryExpr{Left: $1, Op: "&&", Right: $3}
          b.Span = $1.GetSpan().Merge($3.GetSpan())
          $$ = b
        }
    | equality_expr
        { $$ = $1 }
    ;

equality_expr
    : equality_expr EQ relational_expr
        { 
          b := &ast.BinaryExpr{Left: $1, Op: "==", Right: $3}
          b.Span = $1.GetSpan().Merge($3.GetSpan())
          $$ = b
        }
    | equality_expr NEQ relational_expr
        { 
          b := &ast.BinaryExpr{Left: $1, Op: "!=", Right: $3}
          b.Span = $1.GetSpan().Merge($3.GetSpan())
          $$ = b
        }
    | relational_expr
        { $$ = $1 }
    ;

relational_expr
    : relational_expr GT membership_expr
        { 
          b := &ast.BinaryExpr{Left: $1, Op: ">", Right: $3}
          b.Span = $1.GetSpan().Merge($3.GetSpan())
          $$ = b
        }
    | relational_expr LT membership_expr
        { 
          b := &ast.BinaryExpr{Left: $1, Op: "<", Right: $3}
          b.Span = $1.GetSpan().Merge($3.GetSpan())
          $$ = b
        }
    | relational_expr GTE membership_expr
        { 
          b := &ast.BinaryExpr{Left: $1, Op: ">=", Right: $3}
          b.Span = $1.GetSpan().Merge($3.GetSpan())
          $$ = b
        }
    | relational_expr LTE membership_expr
        { 
          b := &ast.BinaryExpr{Left: $1, Op: "<=", Right: $3}
          b.Span = $1.GetSpan().Merge($3.GetSpan())
          $$ = b
        }
    | membership_expr
        { $$ = $1 }
    ;

membership_expr
    : membership_expr IN postfix_expr
        { 
          b := &ast.BinaryExpr{Left: $1, Op: "in", Right: $3}
          b.Span = $1.GetSpan().Merge($3.GetSpan())
          $$ = b
        }
    | additive_expr
        { $$ = $1 }
    ;

additive_expr
    : additive_expr PLUS multiplicative_expr
        { 
          b := &ast.BinaryExpr{Left: $1, Op: "+", Right: $3}
          b.Span = $1.GetSpan().Merge($3.GetSpan())
          $$ = b
        }
    | additive_expr MINUS multiplicative_expr
        { 
          b := &ast.BinaryExpr{Left: $1, Op: "-", Right: $3}
          b.Span = $1.GetSpan().Merge($3.GetSpan())
          $$ = b
        }
    | multiplicative_expr
        { $$ = $1 }
    ;

multiplicative_expr
    : multiplicative_expr STAR unary_expr
        { 
          b := &ast.BinaryExpr{Left: $1, Op: "*", Right: $3}
          b.Span = $1.GetSpan().Merge($3.GetSpan())
          $$ = b
        }
    | multiplicative_expr SLASH unary_expr
        { 
          b := &ast.BinaryExpr{Left: $1, Op: "/", Right: $3}
          b.Span = $1.GetSpan().Merge($3.GetSpan())
          $$ = b
        }
    | multiplicative_expr MOD unary_expr
        { 
            b := &ast.BinaryExpr{Left: $1, Op: "%", Right: $3}
            b.Span = $1.GetSpan().Merge($3.GetSpan())
            $$ = b
        }
    | unary_expr
        { $$ = $1 }
    ;

unary_expr
    : postfix_expr
        { $$ = $1 }
    | NOT unary_expr
        { 
          u := &ast.UnaryExpr{Op: "!", Expr: $2}
          u.Span = $1.Merge($2.GetSpan())
          $$ = u
        }
    | MINUS unary_expr
        { 
          u := &ast.UnaryExpr{Op: "-", Expr: $2}
          u.Span = $1.Merge($2.GetSpan())
          $$ = u
        }
    ;

postfix_expr
    : primary_expr
        { $$ = $1 }
    | IDENT NAMESPACE IDENT LPAREN args_opt RPAREN
        {
          c := &ast.CallExpr{Namespace: $1, Name: $3, Args: $5}
          c.Span = $<span>1.Merge($6)
          $$ = c
        }
    | IDENT LPAREN args_opt RPAREN
        {
          c := &ast.CallExpr{Name: $1, Args: $3}
          c.Span = $<span>1.Merge($4)
          $$ = c
        }
    | postfix_expr DOT IDENT
        { 
          s := &ast.SelectorExpr{Expr: $1, Field: $3}
          s.Span = $1.GetSpan().Merge($<span>3)
          $$ = s
        }
    | postfix_expr LBRACKET expr RBRACKET
        { 
          i := &ast.IndexExpr{Expr: $1, Index: $3}
          i.Span = $1.GetSpan().Merge($4)
          $$ = i
        }
    ;

primary_expr
    : IDENT
        { 
          i := &ast.Identifier{Name: $1}
          i.Span = $<span>1
          $$ = i
        }
    | NUMBER
        { 
          n := &ast.NumberLiteral{Raw: $1}
          n.Span = $<span>1
          $$ = n
        }
    | STRING
        { 
          s := &ast.StringLiteral{Value: $1}
          s.Span = $<span>1
          $$ = s
        }
    | TRUE
        { 
          b := &ast.BoolLiteral{Value: true}
          b.Span = $1
          $$ = b
        }
    | FALSE
        { 
          b := &ast.BoolLiteral{Value: false}
          b.Span = $1
          $$ = b
        }
    | LBRACKET args_opt RBRACKET
        { 
          a := &ast.ArrayLiteral{Elements: $2}
          a.Span = $1.Merge($3)
          $$ = a
        }
    | LBRACE obj_prop_list RBRACE
        { ol := &ast.ObjectLiteral{Properties: $2}
          ol.Span = $1.Merge($3)
          $$ = ol
        }
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

obj_prop
    : IDENT COLON expr
        { 
          p := ast.ObjectProperty{Key: $1, Value: $3}
          p.Span = $<span>1.Merge($3.GetSpan())
          $$ = p
        }
    | STRING COLON expr
        {
          p := ast.ObjectProperty{Key: $1, Value: $3}
          p.Span = $<span>1.Merge($3.GetSpan())
          $$ = p
        }
    ;

obj_prop_list
    : obj_prop
        { $$ = []ast.ObjectProperty{$1} }
    | obj_prop_list COMMA obj_prop
        { $$ = append($1, $3) }
    ;

%%

type ParserError struct {
    Pos     *ast.Position `json:"pos"`
	Message string        `json:"message"`
}

func (pe ParserError) Error() string {
	return fmt.Sprintf("%s: %s", pe.Pos.String(), pe.Message)
}

func Parse(input string) (*ast.Program, error) {
    l := newLexer(input)
    p := yyNewParser()

    if p.Parse(l) != 0 {
        return nil, &ParserError{
            Pos: &ast.Position{ Line: l.line, Col: l.col },
			Message: l.lastError,
        }
    }

    return l.result, nil
}