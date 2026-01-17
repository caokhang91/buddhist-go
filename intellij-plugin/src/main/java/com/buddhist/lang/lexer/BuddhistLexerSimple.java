package com.buddhist.lang.lexer;

import com.buddhist.lang.psi.BuddhistTypes;
import com.intellij.lexer.LexerBase;
import com.intellij.psi.tree.IElementType;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.util.HashMap;
import java.util.Map;

public class BuddhistLexerSimple extends LexerBase {
    private CharSequence buffer;
    private int startOffset;
    private int endOffset;
    private int currentOffset;
    private IElementType currentToken;
    private int tokenStart;

    private static final Map<String, IElementType> KEYWORDS = new HashMap<>();
    
    static {
        KEYWORDS.put("fn", BuddhistTypes.FUNCTION);
        KEYWORDS.put("let", BuddhistTypes.LET);
        KEYWORDS.put("const", BuddhistTypes.CONST);
        KEYWORDS.put("true", BuddhistTypes.TRUE);
        KEYWORDS.put("false", BuddhistTypes.FALSE);
        KEYWORDS.put("if", BuddhistTypes.IF);
        KEYWORDS.put("else", BuddhistTypes.ELSE);
        KEYWORDS.put("return", BuddhistTypes.RETURN);
        KEYWORDS.put("for", BuddhistTypes.FOR);
        KEYWORDS.put("while", BuddhistTypes.WHILE);
        KEYWORDS.put("break", BuddhistTypes.BREAK);
        KEYWORDS.put("continue", BuddhistTypes.CONTINUE);
        KEYWORDS.put("null", BuddhistTypes.NULL);
        KEYWORDS.put("spawn", BuddhistTypes.SPAWN);
        KEYWORDS.put("channel", BuddhistTypes.CHANNEL);
        KEYWORDS.put("class", BuddhistTypes.CLASS);
        KEYWORDS.put("import", BuddhistTypes.IMPORT);
        KEYWORDS.put("export", BuddhistTypes.EXPORT);
        KEYWORDS.put("from", BuddhistTypes.FROM);
        KEYWORDS.put("try", BuddhistTypes.TRY);
        KEYWORDS.put("catch", BuddhistTypes.CATCH);
        KEYWORDS.put("finally", BuddhistTypes.FINALLY);
        KEYWORDS.put("throw", BuddhistTypes.THROW);
        KEYWORDS.put("blob", BuddhistTypes.BLOB);
    }

    @Override
    public void start(@NotNull CharSequence buffer, int startOffset, int endOffset, int initialState) {
        this.buffer = buffer;
        this.startOffset = startOffset;
        this.endOffset = endOffset;
        this.currentOffset = startOffset;
        this.tokenStart = startOffset;
        advance();
    }

    @Override
    public int getState() {
        return 0;
    }

    @Nullable
    @Override
    public IElementType getTokenType() {
        return currentToken;
    }

    @Override
    public int getTokenStart() {
        return tokenStart;
    }

    @Override
    public int getTokenEnd() {
        return currentOffset;
    }

    @Override
    public void advance() {
        tokenStart = currentOffset;
        
        if (currentOffset >= endOffset) {
            currentToken = null;
            return;
        }

        char ch = buffer.charAt(currentOffset);

        // Skip whitespace
        if (Character.isWhitespace(ch)) {
            while (currentOffset < endOffset && Character.isWhitespace(buffer.charAt(currentOffset))) {
                currentOffset++;
            }
            currentToken = com.intellij.psi.TokenType.WHITE_SPACE;
            return;
        }

        // Line comment
        if (ch == '/' && currentOffset + 1 < endOffset && buffer.charAt(currentOffset + 1) == '/') {
            while (currentOffset < endOffset && buffer.charAt(currentOffset) != '\n' && buffer.charAt(currentOffset) != '\r') {
                currentOffset++;
            }
            currentToken = BuddhistTypes.LINE_COMMENT;
            return;
        }

        // Block comment
        if (ch == '/' && currentOffset + 1 < endOffset && buffer.charAt(currentOffset + 1) == '*') {
            currentOffset += 2;
            while (currentOffset < endOffset - 1) {
                if (buffer.charAt(currentOffset) == '*' && buffer.charAt(currentOffset + 1) == '/') {
                    currentOffset += 2;
                    break;
                }
                currentOffset++;
            }
            currentToken = BuddhistTypes.BLOCK_COMMENT;
            return;
        }

        // String literal
        if (ch == '"') {
            currentOffset++;
            while (currentOffset < endOffset) {
                char c = buffer.charAt(currentOffset);
                if (c == '"' && (currentOffset == tokenStart + 1 || buffer.charAt(currentOffset - 1) != '\\')) {
                    currentOffset++;
                    break;
                }
                currentOffset++;
            }
            currentToken = BuddhistTypes.STRING;
            return;
        }

        // Numbers
        if (Character.isDigit(ch)) {
            while (currentOffset < endOffset && Character.isDigit(buffer.charAt(currentOffset))) {
                currentOffset++;
            }
            if (currentOffset < endOffset && buffer.charAt(currentOffset) == '.' && 
                currentOffset + 1 < endOffset && Character.isDigit(buffer.charAt(currentOffset + 1))) {
                currentOffset++;
                while (currentOffset < endOffset && Character.isDigit(buffer.charAt(currentOffset))) {
                    currentOffset++;
                }
                currentToken = BuddhistTypes.FLOAT;
            } else {
                currentToken = BuddhistTypes.INT;
            }
            return;
        }

        // Identifiers and keywords
        if (Character.isLetter(ch) || ch == '_') {
            int start = currentOffset;
            while (currentOffset < endOffset && 
                   (Character.isLetterOrDigit(buffer.charAt(currentOffset)) || buffer.charAt(currentOffset) == '_')) {
                currentOffset++;
            }
            String identifier = buffer.subSequence(start, currentOffset).toString();
            currentToken = KEYWORDS.getOrDefault(identifier, BuddhistTypes.IDENTIFIER);
            return;
        }

        // Operators and delimiters
        switch (ch) {
            case '=':
                if (currentOffset + 1 < endOffset && buffer.charAt(currentOffset + 1) == '=') {
                    currentOffset += 2;
                    currentToken = BuddhistTypes.EQ;
                } else if (currentOffset + 1 < endOffset && buffer.charAt(currentOffset + 1) == '>') {
                    currentOffset += 2;
                    currentToken = BuddhistTypes.ARROW;
                } else {
                    currentOffset++;
                    currentToken = BuddhistTypes.ASSIGN;
                }
                return;
            case '!':
                if (currentOffset + 1 < endOffset && buffer.charAt(currentOffset + 1) == '=') {
                    currentOffset += 2;
                    currentToken = BuddhistTypes.NOT_EQ;
                } else {
                    currentOffset++;
                    currentToken = BuddhistTypes.BANG;
                }
                return;
            case '<':
                if (currentOffset + 1 < endOffset && buffer.charAt(currentOffset + 1) == '=') {
                    currentOffset += 2;
                    currentToken = BuddhistTypes.LT_EQ;
                } else if (currentOffset + 1 < endOffset && buffer.charAt(currentOffset + 1) == '-') {
                    currentOffset += 2;
                    currentToken = BuddhistTypes.SEND;
                } else {
                    currentOffset++;
                    currentToken = BuddhistTypes.LT;
                }
                return;
            case '>':
                if (currentOffset + 1 < endOffset && buffer.charAt(currentOffset + 1) == '=') {
                    currentOffset += 2;
                    currentToken = BuddhistTypes.GT_EQ;
                } else if (currentOffset + 1 < endOffset && buffer.charAt(currentOffset + 1) == '-') {
                    currentOffset += 2;
                    currentToken = BuddhistTypes.RECEIVE;
                } else {
                    currentOffset++;
                    currentToken = BuddhistTypes.GT;
                }
                return;
            case '&':
                if (currentOffset + 1 < endOffset && buffer.charAt(currentOffset + 1) == '&') {
                    currentOffset += 2;
                    currentToken = BuddhistTypes.AND;
                } else {
                    currentOffset++;
                    currentToken = BuddhistTypes.BAD_CHARACTER;
                }
                return;
            case '|':
                if (currentOffset + 1 < endOffset && buffer.charAt(currentOffset + 1) == '|') {
                    currentOffset += 2;
                    currentToken = BuddhistTypes.OR;
                } else {
                    currentOffset++;
                    currentToken = BuddhistTypes.BAD_CHARACTER;
                }
                return;
            case '+':
                currentOffset++;
                currentToken = BuddhistTypes.PLUS;
                return;
            case '-':
                currentOffset++;
                currentToken = BuddhistTypes.MINUS;
                return;
            case '*':
                currentOffset++;
                currentToken = BuddhistTypes.ASTERISK;
                return;
            case '/':
                currentOffset++;
                currentToken = BuddhistTypes.SLASH;
                return;
            case '%':
                currentOffset++;
                currentToken = BuddhistTypes.MODULO;
                return;
            case ',':
                currentOffset++;
                currentToken = BuddhistTypes.COMMA;
                return;
            case ';':
                currentOffset++;
                currentToken = BuddhistTypes.SEMICOLON;
                return;
            case ':':
                currentOffset++;
                currentToken = BuddhistTypes.COLON;
                return;
            case '(':
                currentOffset++;
                currentToken = BuddhistTypes.LPAREN;
                return;
            case ')':
                currentOffset++;
                currentToken = BuddhistTypes.RPAREN;
                return;
            case '{':
                currentOffset++;
                currentToken = BuddhistTypes.LBRACE;
                return;
            case '}':
                currentOffset++;
                currentToken = BuddhistTypes.RBRACE;
                return;
            case '[':
                currentOffset++;
                currentToken = BuddhistTypes.LBRACKET;
                return;
            case ']':
                currentOffset++;
                currentToken = BuddhistTypes.RBRACKET;
                return;
            default:
                currentOffset++;
                currentToken = BuddhistTypes.BAD_CHARACTER;
        }
    }

    @NotNull
    @Override
    public CharSequence getBufferSequence() {
        return buffer;
    }

    @Override
    public int getBufferEnd() {
        return endOffset;
    }
}
