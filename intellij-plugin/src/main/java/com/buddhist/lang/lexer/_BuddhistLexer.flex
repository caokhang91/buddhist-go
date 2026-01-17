package com.buddhist.lang.lexer;

import com.intellij.lexer.FlexLexer;
import com.intellij.psi.tree.IElementType;
import com.buddhist.lang.psi.BuddhistTypes;

%%

%class _BuddhistLexer
%implements FlexLexer
%unicode
%function advance
%type IElementType
%eof{  return;
%eof}

%{
  private int commentStart;
  private int braceCount;
%}

EOL=\r|\n|\r\n
WHITE_SPACE=[\ \t\f]+

// Comments
LINE_COMMENT="//".*
BLOCK_COMMENT_START="/*"
BLOCK_COMMENT_END="*/"

// Identifiers
IDENTIFIER=[a-zA-Z_][a-zA-Z0-9_]*

// Numbers
INTEGER=[0-9]+
FLOAT=[0-9]+\.[0-9]+

// Strings
STRING_LITERAL=\"([^\"\\]|\\.)*\"

// Operators
ASSIGN="="
PLUS="+"
MINUS="-"
BANG="!"
ASTERISK="*"
SLASH="/"
MODULO="%"
LT="<"
GT=">"
EQ="=="
NOT_EQ="!="
LT_EQ="<="
GT_EQ=">="
AND="&&"
OR="||"
ARROW="=>"
SEND="<-"
RECEIVE="->"

// Delimiters
COMMA=","
SEMICOLON=";"
COLON=":"
LPAREN="("
RPAREN=")"
LBRACE="{"
RBRACE="}"
LBRACKET="["
RBRACKET="]"

%state IN_BLOCK_COMMENT

%%

<YYINITIAL> {
    {WHITE_SPACE}              { return com.intellij.psi.TokenType.WHITE_SPACE; }
    {EOL}                      { return com.intellij.psi.TokenType.WHITE_SPACE; }

    {LINE_COMMENT}             { return BuddhistTypes.LINE_COMMENT; }
    {BLOCK_COMMENT_START}      { yybegin(IN_BLOCK_COMMENT); commentStart = zzStartRead; return BuddhistTypes.BLOCK_COMMENT_START; }

    // Keywords
    "fn"                       { return BuddhistTypes.FUNCTION; }
    "let"                      { return BuddhistTypes.LET; }
    "const"                    { return BuddhistTypes.CONST; }
    "true"                     { return BuddhistTypes.TRUE; }
    "false"                    { return BuddhistTypes.FALSE; }
    "if"                       { return BuddhistTypes.IF; }
    "else"                     { return BuddhistTypes.ELSE; }
    "return"                   { return BuddhistTypes.RETURN; }
    "for"                      { return BuddhistTypes.FOR; }
    "while"                    { return BuddhistTypes.WHILE; }
    "break"                    { return BuddhistTypes.BREAK; }
    "continue"                 { return BuddhistTypes.CONTINUE; }
    "null"                     { return BuddhistTypes.NULL; }
    "spawn"                    { return BuddhistTypes.SPAWN; }
    "channel"                  { return BuddhistTypes.CHANNEL; }
    "class"                    { return BuddhistTypes.CLASS; }
    "import"                   { return BuddhistTypes.IMPORT; }
    "export"                   { return BuddhistTypes.EXPORT; }
    "from"                     { return BuddhistTypes.FROM; }
    "try"                      { return BuddhistTypes.TRY; }
    "catch"                    { return BuddhistTypes.CATCH; }
    "finally"                  { return BuddhistTypes.FINALLY; }
    "throw"                    { return BuddhistTypes.THROW; }
    "blob"                     { return BuddhistTypes.BLOB; }

    // Literals
    {INTEGER}                  { return BuddhistTypes.INT; }
    {FLOAT}                    { return BuddhistTypes.FLOAT; }
    {STRING_LITERAL}           { return BuddhistTypes.STRING; }
    {IDENTIFIER}               { return BuddhistTypes.IDENTIFIER; }

    // Operators
    {ASSIGN}                   { return BuddhistTypes.ASSIGN; }
    {PLUS}                     { return BuddhistTypes.PLUS; }
    {MINUS}                    { return BuddhistTypes.MINUS; }
    {BANG}                     { return BuddhistTypes.BANG; }
    {ASTERISK}                 { return BuddhistTypes.ASTERISK; }
    {SLASH}                    { return BuddhistTypes.SLASH; }
    {MODULO}                   { return BuddhistTypes.MODULO; }
    {LT}                       { return BuddhistTypes.LT; }
    {GT}                       { return BuddhistTypes.GT; }
    {EQ}                       { return BuddhistTypes.EQ; }
    {NOT_EQ}                   { return BuddhistTypes.NOT_EQ; }
    {LT_EQ}                    { return BuddhistTypes.LT_EQ; }
    {GT_EQ}                    { return BuddhistTypes.GT_EQ; }
    {AND}                      { return BuddhistTypes.AND; }
    {OR}                       { return BuddhistTypes.OR; }
    {ARROW}                    { return BuddhistTypes.ARROW; }
    {SEND}                     { return BuddhistTypes.SEND; }
    {RECEIVE}                  { return BuddhistTypes.RECEIVE; }

    // Delimiters
    {COMMA}                    { return BuddhistTypes.COMMA; }
    {SEMICOLON}                { return BuddhistTypes.SEMICOLON; }
    {COLON}                    { return BuddhistTypes.COLON; }
    {LPAREN}                   { return BuddhistTypes.LPAREN; }
    {RPAREN}                   { return BuddhistTypes.RPAREN; }
    {LBRACE}                   { return BuddhistTypes.LBRACE; }
    {RBRACE}                   { return BuddhistTypes.RBRACE; }
    {LBRACKET}                 { return BuddhistTypes.LBRACKET; }
    {RBRACKET}                 { return BuddhistTypes.RBRACKET; }
}

<IN_BLOCK_COMMENT> {
    {BLOCK_COMMENT_END}        { yybegin(YYINITIAL); return BuddhistTypes.BLOCK_COMMENT_END; }
    [^]                        { return BuddhistTypes.BLOCK_COMMENT; }
}

[^]                            { return BuddhistTypes.BAD_CHARACTER; }
