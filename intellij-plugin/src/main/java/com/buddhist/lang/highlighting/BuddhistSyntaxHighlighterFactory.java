package com.buddhist.lang.highlighting;

import com.buddhist.lang.BuddhistLanguage;
import com.buddhist.lang.lexer.BuddhistLexer;
import com.intellij.lexer.Lexer;
import com.intellij.openapi.editor.DefaultLanguageHighlighterColors;
import com.intellij.openapi.editor.HighlighterColors;
import com.intellij.openapi.editor.colors.TextAttributesKey;
import com.intellij.openapi.fileTypes.SyntaxHighlighter;
import com.intellij.openapi.fileTypes.SyntaxHighlighterFactory;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.psi.tree.IElementType;
import com.buddhist.lang.psi.BuddhistTypes;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class BuddhistSyntaxHighlighterFactory extends SyntaxHighlighterFactory {
    @NotNull
    @Override
    public SyntaxHighlighter getSyntaxHighlighter(@Nullable Project project, @Nullable VirtualFile virtualFile) {
        return new BuddhistSyntaxHighlighter();
    }

    private static class BuddhistSyntaxHighlighter implements SyntaxHighlighter {
        private static final TextAttributesKey[] EMPTY_KEYS = new TextAttributesKey[0];

        @NotNull
        @Override
        public Lexer getHighlightingLexer() {
            return BuddhistLexer.createLexer();
        }

        @NotNull
        @Override
        public TextAttributesKey[] getTokenHighlights(IElementType tokenType) {
            if (tokenType.equals(BuddhistTypes.LINE_COMMENT) || 
                tokenType.equals(BuddhistTypes.BLOCK_COMMENT)) {
                return new TextAttributesKey[]{BuddhistSyntaxHighlighterColors.COMMENT};
            }
            if (tokenType.equals(BuddhistTypes.STRING)) {
                return new TextAttributesKey[]{BuddhistSyntaxHighlighterColors.STRING};
            }
            if (tokenType.equals(BuddhistTypes.INT) || tokenType.equals(BuddhistTypes.FLOAT)) {
                return new TextAttributesKey[]{BuddhistSyntaxHighlighterColors.NUMBER};
            }
            if (isKeyword(tokenType)) {
                return new TextAttributesKey[]{BuddhistSyntaxHighlighterColors.KEYWORD};
            }
            if (tokenType.equals(BuddhistTypes.IDENTIFIER)) {
                return new TextAttributesKey[]{BuddhistSyntaxHighlighterColors.IDENTIFIER};
            }
            if (isOperator(tokenType)) {
                return new TextAttributesKey[]{BuddhistSyntaxHighlighterColors.OPERATOR};
            }
            return EMPTY_KEYS;
        }

        private boolean isKeyword(IElementType tokenType) {
            return tokenType.equals(BuddhistTypes.FUNCTION) ||
                   tokenType.equals(BuddhistTypes.LET) ||
                   tokenType.equals(BuddhistTypes.CONST) ||
                   tokenType.equals(BuddhistTypes.IF) ||
                   tokenType.equals(BuddhistTypes.ELSE) ||
                   tokenType.equals(BuddhistTypes.RETURN) ||
                   tokenType.equals(BuddhistTypes.FOR) ||
                   tokenType.equals(BuddhistTypes.WHILE) ||
                   tokenType.equals(BuddhistTypes.BREAK) ||
                   tokenType.equals(BuddhistTypes.CONTINUE) ||
                   tokenType.equals(BuddhistTypes.TRUE) ||
                   tokenType.equals(BuddhistTypes.FALSE) ||
                   tokenType.equals(BuddhistTypes.NULL) ||
                   tokenType.equals(BuddhistTypes.SPAWN) ||
                   tokenType.equals(BuddhistTypes.CHANNEL) ||
                   tokenType.equals(BuddhistTypes.CLASS) ||
                   tokenType.equals(BuddhistTypes.IMPORT) ||
                   tokenType.equals(BuddhistTypes.EXPORT) ||
                   tokenType.equals(BuddhistTypes.FROM) ||
                   tokenType.equals(BuddhistTypes.TRY) ||
                   tokenType.equals(BuddhistTypes.CATCH) ||
                   tokenType.equals(BuddhistTypes.FINALLY) ||
                   tokenType.equals(BuddhistTypes.THROW) ||
                   tokenType.equals(BuddhistTypes.BLOB);
        }

        private boolean isOperator(IElementType tokenType) {
            return tokenType.equals(BuddhistTypes.PLUS) ||
                   tokenType.equals(BuddhistTypes.MINUS) ||
                   tokenType.equals(BuddhistTypes.ASTERISK) ||
                   tokenType.equals(BuddhistTypes.SLASH) ||
                   tokenType.equals(BuddhistTypes.MODULO) ||
                   tokenType.equals(BuddhistTypes.ASSIGN) ||
                   tokenType.equals(BuddhistTypes.EQ) ||
                   tokenType.equals(BuddhistTypes.NOT_EQ) ||
                   tokenType.equals(BuddhistTypes.LT) ||
                   tokenType.equals(BuddhistTypes.GT) ||
                   tokenType.equals(BuddhistTypes.LT_EQ) ||
                   tokenType.equals(BuddhistTypes.GT_EQ) ||
                   tokenType.equals(BuddhistTypes.AND) ||
                   tokenType.equals(BuddhistTypes.OR) ||
                   tokenType.equals(BuddhistTypes.ARROW) ||
                   tokenType.equals(BuddhistTypes.SEND) ||
                   tokenType.equals(BuddhistTypes.RECEIVE);
        }
    }
}
