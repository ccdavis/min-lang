# MinLang Grammar (BNF)

## Program Structure

```bnf
<program>         ::= <declaration>*

<declaration>     ::= <var-decl> | <const-decl> | <func-decl> | <struct-decl>
```

## Declarations

```bnf
<var-decl>        ::= "var" <identifier> <type-annotation>? "=" <expression> ";"
                    | "var" <identifier> <type-annotation> ";"

<const-decl>      ::= "const" <identifier> <type-annotation>? "=" <expression> ";"

<func-decl>       ::= "func" <identifier> "(" <param-list>? ")" <type-annotation>? <block>

<struct-decl>     ::= "struct" <identifier> "{" <field-list> "}"

<param-list>      ::= <param> ("," <param>)*

<param>           ::= <identifier> <type-annotation>

<field-list>      ::= <field> (";" <field>)* ";"?

<field>           ::= <identifier> <type-annotation>
```

## Types

```bnf
<type-annotation> ::= ":" <type>

<type>            ::= <identifier>                        # Named type (int, float, bool, string, or struct)
                    | "[" "]" <type>                      # Array type
                    | "map" "[" <type> "]" <type>         # Map type
                    | "func" "(" <type-list>? ")" <type>? # Function type

<type-list>       ::= <type> ("," <type>)*
```

## Statements

```bnf
<statement>       ::= <var-decl>
                    | <const-decl>
                    | <func-decl>
                    | <assignment>
                    | <if-stmt>
                    | <for-stmt>
                    | <return-stmt>
                    | <expr-stmt>
                    | <block>

<assignment>      ::= <identifier> ("." <identifier> | "[" <expression> "]")* "=" <expression> ";"

<if-stmt>         ::= "if" <expression> <block> ("else" (<if-stmt> | <block>))?

<for-stmt>        ::= "for" <expression> <block>
                    | "for" <var-decl> <expression> ";" <assignment> <block>

<return-stmt>     ::= "return" <expression>? ";"

<expr-stmt>       ::= <expression> ";"

<block>           ::= "{" <statement>* "}"
```

## Expressions

```bnf
<expression>      ::= <assignment-expr>

<assignment-expr> ::= <logical-or>

<logical-or>      ::= <logical-and> ("||" <logical-and>)*

<logical-and>     ::= <equality> ("&&" <equality>)*

<equality>        ::= <comparison> (("==" | "!=") <comparison>)*

<comparison>      ::= <additive> (("<" | ">" | "<=" | ">=") <additive>)*

<additive>        ::= <multiplicative> (("+" | "-") <multiplicative>)*

<multiplicative>  ::= <unary> (("*" | "/" | "%") <unary>)*

<unary>           ::= ("!" | "-") <unary>
                    | <postfix>

<postfix>         ::= <primary> <postfix-op>*

<postfix-op>      ::= "(" <arg-list>? ")"              # Function call
                    | "[" <expression> "]"             # Array/Map index
                    | "." <identifier>                 # Field access

<primary>         ::= <identifier>
                    | <integer>
                    | <float>
                    | <string>
                    | <boolean>
                    | "nil"
                    | <array-literal>
                    | <map-literal>
                    | <struct-literal>
                    | "(" <expression> ")"

<arg-list>        ::= <expression> ("," <expression>)*

<array-literal>   ::= "[" <arg-list>? "]"

<map-literal>     ::= "map" "[" <type> "]" <type> "{" <map-entry-list>? "}"

<map-entry-list>  ::= <map-entry> ("," <map-entry>)* ","?

<map-entry>       ::= <expression> ":" <expression>

<struct-literal>  ::= <identifier> "{" <field-init-list>? "}"

<field-init-list> ::= <field-init> ("," <field-init>)* ","?

<field-init>      ::= <identifier> ":" <expression>
```

## Lexical Elements

```bnf
<identifier>      ::= [a-zA-Z_][a-zA-Z0-9_]*

<integer>         ::= [0-9]+

<float>           ::= [0-9]+ "." [0-9]+

<string>          ::= '"' ([^"\\] | '\\' .)* '"'

<boolean>         ::= "true" | "false"
```

## Keywords

```
var const func struct return if else for map true false nil
```

## Operators and Delimiters

```
+ - * / % == != < > <= >= && || ! = : ; , . ( ) { } [ ]
```

## Comments

```bnf
<line-comment>    ::= "//" [^\n]* "\n"

<block-comment>   ::= "/*" .* "*/"
```
