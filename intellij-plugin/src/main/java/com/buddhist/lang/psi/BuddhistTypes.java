package com.buddhist.lang.psi;

import com.intellij.psi.tree.IElementType;
import com.buddhist.lang.BuddhistLanguage;
import org.jetbrains.annotations.NonNls;
import org.jetbrains.annotations.NotNull;

public class BuddhistTypes {
    // Token types
    public static final IElementType LINE_COMMENT = new BuddhistElementType("LINE_COMMENT");
    public static final IElementType BLOCK_COMMENT = new BuddhistElementType("BLOCK_COMMENT");
    public static final IElementType BLOCK_COMMENT_START = new BuddhistElementType("BLOCK_COMMENT_START");
    public static final IElementType BLOCK_COMMENT_END = new BuddhistElementType("BLOCK_COMMENT_END");

    // Keywords
    public static final IElementType FUNCTION = new BuddhistElementType("FUNCTION");
    public static final IElementType LET = new BuddhistElementType("LET");
    public static final IElementType CONST = new BuddhistElementType("CONST");
    public static final IElementType TRUE = new BuddhistElementType("TRUE");
    public static final IElementType FALSE = new BuddhistElementType("FALSE");
    public static final IElementType IF = new BuddhistElementType("IF");
    public static final IElementType ELSE = new BuddhistElementType("ELSE");
    public static final IElementType RETURN = new BuddhistElementType("RETURN");
    public static final IElementType FOR = new BuddhistElementType("FOR");
    public static final IElementType WHILE = new BuddhistElementType("WHILE");
    public static final IElementType BREAK = new BuddhistElementType("BREAK");
    public static final IElementType CONTINUE = new BuddhistElementType("CONTINUE");
    public static final IElementType NULL = new BuddhistElementType("NULL");
    public static final IElementType SPAWN = new BuddhistElementType("SPAWN");
    public static final IElementType CHANNEL = new BuddhistElementType("CHANNEL");
    public static final IElementType CLASS = new BuddhistElementType("CLASS");
    public static final IElementType IMPORT = new BuddhistElementType("IMPORT");
    public static final IElementType EXPORT = new BuddhistElementType("EXPORT");
    public static final IElementType FROM = new BuddhistElementType("FROM");
    public static final IElementType TRY = new BuddhistElementType("TRY");
    public static final IElementType CATCH = new BuddhistElementType("CATCH");
    public static final IElementType FINALLY = new BuddhistElementType("FINALLY");
    public static final IElementType THROW = new BuddhistElementType("THROW");
    public static final IElementType BLOB = new BuddhistElementType("BLOB");

    // Literals
    public static final IElementType IDENTIFIER = new BuddhistElementType("IDENTIFIER");
    public static final IElementType INT = new BuddhistElementType("INT");
    public static final IElementType FLOAT = new BuddhistElementType("FLOAT");
    public static final IElementType STRING = new BuddhistElementType("STRING");

    // Operators
    public static final IElementType ASSIGN = new BuddhistElementType("ASSIGN");
    public static final IElementType PLUS = new BuddhistElementType("PLUS");
    public static final IElementType MINUS = new BuddhistElementType("MINUS");
    public static final IElementType BANG = new BuddhistElementType("BANG");
    public static final IElementType ASTERISK = new BuddhistElementType("ASTERISK");
    public static final IElementType SLASH = new BuddhistElementType("SLASH");
    public static final IElementType MODULO = new BuddhistElementType("MODULO");
    public static final IElementType LT = new BuddhistElementType("LT");
    public static final IElementType GT = new BuddhistElementType("GT");
    public static final IElementType EQ = new BuddhistElementType("EQ");
    public static final IElementType NOT_EQ = new BuddhistElementType("NOT_EQ");
    public static final IElementType LT_EQ = new BuddhistElementType("LT_EQ");
    public static final IElementType GT_EQ = new BuddhistElementType("GT_EQ");
    public static final IElementType AND = new BuddhistElementType("AND");
    public static final IElementType OR = new BuddhistElementType("OR");
    public static final IElementType ARROW = new BuddhistElementType("ARROW");
    public static final IElementType SEND = new BuddhistElementType("SEND");
    public static final IElementType RECEIVE = new BuddhistElementType("RECEIVE");

    // Delimiters
    public static final IElementType COMMA = new BuddhistElementType("COMMA");
    public static final IElementType SEMICOLON = new BuddhistElementType("SEMICOLON");
    public static final IElementType COLON = new BuddhistElementType("COLON");
    public static final IElementType LPAREN = new BuddhistElementType("LPAREN");
    public static final IElementType RPAREN = new BuddhistElementType("RPAREN");
    public static final IElementType LBRACE = new BuddhistElementType("LBRACE");
    public static final IElementType RBRACE = new BuddhistElementType("RBRACE");
    public static final IElementType LBRACKET = new BuddhistElementType("LBRACKET");
    public static final IElementType RBRACKET = new BuddhistElementType("RBRACKET");

    // Bad character
    public static final IElementType BAD_CHARACTER = new BuddhistElementType("BAD_CHARACTER");

    private static class BuddhistElementType extends IElementType {
        public BuddhistElementType(@NotNull @NonNls String debugName) {
            super(debugName, BuddhistLanguage.INSTANCE);
        }
    }
}
