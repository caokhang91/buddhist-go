package com.buddhist.lang.parser;

import com.buddhist.lang.BuddhistLanguage;
import com.buddhist.lang.lexer.BuddhistLexer;
import com.intellij.lang.ASTNode;
import com.intellij.lang.ParserDefinition;
import com.intellij.lang.PsiParser;
import com.intellij.lexer.Lexer;
import com.intellij.openapi.project.Project;
import com.intellij.psi.FileViewProvider;
import com.intellij.psi.PsiElement;
import com.intellij.psi.PsiFile;
import com.intellij.psi.tree.IFileElementType;
import com.intellij.psi.tree.TokenSet;
import com.buddhist.lang.psi.BuddhistTypes;
import com.buddhist.lang.psi.BuddhistFile;
import org.jetbrains.annotations.NotNull;

public class BuddhistParserDefinition implements ParserDefinition {
    public static final IFileElementType FILE = new IFileElementType(BuddhistLanguage.INSTANCE);

    public static final TokenSet WHITE_SPACES = TokenSet.create(com.intellij.psi.TokenType.WHITE_SPACE);
    public static final TokenSet COMMENTS = TokenSet.create(
            BuddhistTypes.LINE_COMMENT,
            BuddhistTypes.BLOCK_COMMENT
    );
    public static final TokenSet STRING_LITERALS = TokenSet.create(BuddhistTypes.STRING);

    @NotNull
    @Override
    public Lexer createLexer(Project project) {
        return BuddhistLexer.createLexer();
    }

    @NotNull
    @Override
    public TokenSet getWhitespaceTokens() {
        return WHITE_SPACES;
    }

    @NotNull
    @Override
    public TokenSet getCommentTokens() {
        return COMMENTS;
    }

    @NotNull
    @Override
    public TokenSet getStringLiteralElements() {
        return STRING_LITERALS;
    }

    @NotNull
    @Override
    public PsiParser createParser(final Project project) {
        return new BuddhistParser();
    }

    @NotNull
    @Override
    public IFileElementType getFileNodeType() {
        return FILE;
    }

    @NotNull
    @Override
    public PsiFile createFile(@NotNull FileViewProvider viewProvider) {
        return new BuddhistFile(viewProvider);
    }

    @NotNull
    @Override
    public SpaceRequirements spaceExistenceTypeBetweenTokens(ASTNode left, ASTNode right) {
        return SpaceRequirements.MAY;
    }

    @NotNull
    @Override
    public PsiElement createElement(ASTNode node) {
        return com.intellij.psi.impl.source.tree.LeafPsiElement.Factory.createElement(node);
    }
}
