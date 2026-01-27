package com.buddhist.lang.parser;

import com.buddhist.lang.psi.BuddhistTypes;
import com.intellij.lang.ASTNode;
import com.intellij.lang.PsiBuilder;
import com.intellij.lang.PsiBuilderFactory;
import com.intellij.lang.PsiParser;
import com.intellij.psi.tree.IElementType;
import org.jetbrains.annotations.NotNull;

public class BuddhistParser implements PsiParser {
    @NotNull
    @Override
    public ASTNode parse(@NotNull IElementType root, @NotNull com.intellij.lang.PsiBuilder builder) {
        PsiBuilder.Marker rootMarker = builder.mark();
        
        while (!builder.eof()) {
            parseStatement(builder);
        }
        
        rootMarker.done(root);
        return builder.getTreeBuilt();
    }

    private void parseStatement(PsiBuilder builder) {
        IElementType tokenType = builder.getTokenType();
        
        if (tokenType == BuddhistTypes.LET) {
            parseLetStatement(builder);
        } else if (tokenType == BuddhistTypes.CONST) {
            parseConstStatement(builder);
        } else if (tokenType == BuddhistTypes.RETURN) {
            parseReturnStatement(builder);
        } else if (tokenType == BuddhistTypes.IF) {
            parseIfStatement(builder);
        } else if (tokenType == BuddhistTypes.WHILE) {
            parseWhileStatement(builder);
        } else if (tokenType == BuddhistTypes.FOR) {
            parseForStatement(builder);
        } else if (tokenType == BuddhistTypes.FUNCTION) {
            parseFunctionStatement(builder);
        } else {
            parseExpression(builder);
            if (builder.getTokenType() == BuddhistTypes.SEMICOLON) {
                builder.advanceLexer();
            }
        }
    }

    private void parseLetStatement(PsiBuilder builder) {
        PsiBuilder.Marker marker = builder.mark();
        builder.advanceLexer(); // LET
        if (builder.getTokenType() == BuddhistTypes.IDENTIFIER) {
            builder.advanceLexer(); // identifier
        }
        if (builder.getTokenType() == BuddhistTypes.ASSIGN) {
            builder.advanceLexer(); // =
            parseExpression(builder);
        }
        if (builder.getTokenType() == BuddhistTypes.SEMICOLON) {
            builder.advanceLexer(); // ;
        }
        marker.done(BuddhistTypes.LET);
    }

    private void parseConstStatement(PsiBuilder builder) {
        PsiBuilder.Marker marker = builder.mark();
        builder.advanceLexer(); // CONST
        if (builder.getTokenType() == BuddhistTypes.IDENTIFIER) {
            builder.advanceLexer(); // identifier
        }
        if (builder.getTokenType() == BuddhistTypes.ASSIGN) {
            builder.advanceLexer(); // =
            parseExpression(builder);
        }
        if (builder.getTokenType() == BuddhistTypes.SEMICOLON) {
            builder.advanceLexer(); // ;
        }
        marker.done(BuddhistTypes.CONST);
    }

    private void parseReturnStatement(PsiBuilder builder) {
        PsiBuilder.Marker marker = builder.mark();
        builder.advanceLexer(); // RETURN
        parseExpression(builder);
        if (builder.getTokenType() == BuddhistTypes.SEMICOLON) {
            builder.advanceLexer(); // ;
        }
        marker.done(BuddhistTypes.RETURN);
    }

    private void parseIfStatement(PsiBuilder builder) {
        PsiBuilder.Marker marker = builder.mark();
        builder.advanceLexer(); // IF
        if (builder.getTokenType() == BuddhistTypes.LPAREN) {
            builder.advanceLexer(); // (
            parseExpression(builder);
            if (builder.getTokenType() == BuddhistTypes.RPAREN) {
                builder.advanceLexer(); // )
            }
        }
        if (builder.getTokenType() == BuddhistTypes.LBRACE) {
            parseBlock(builder);
        }
        if (builder.getTokenType() == BuddhistTypes.ELSE) {
            builder.advanceLexer(); // ELSE
            if (builder.getTokenType() == BuddhistTypes.LBRACE) {
                parseBlock(builder);
            }
        }
        marker.done(BuddhistTypes.IF);
    }

    private void parseWhileStatement(PsiBuilder builder) {
        PsiBuilder.Marker marker = builder.mark();
        builder.advanceLexer(); // WHILE
        if (builder.getTokenType() == BuddhistTypes.LPAREN) {
            builder.advanceLexer(); // (
            parseExpression(builder);
            if (builder.getTokenType() == BuddhistTypes.RPAREN) {
                builder.advanceLexer(); // )
            }
        }
        if (builder.getTokenType() == BuddhistTypes.LBRACE) {
            parseBlock(builder);
        }
        marker.done(BuddhistTypes.WHILE);
    }

    private void parseForStatement(PsiBuilder builder) {
        PsiBuilder.Marker marker = builder.mark();
        builder.advanceLexer(); // FOR
        if (builder.getTokenType() == BuddhistTypes.LPAREN) {
            builder.advanceLexer(); // (
            parseStatement(builder); // init
            parseExpression(builder); // condition
            if (builder.getTokenType() == BuddhistTypes.SEMICOLON) {
                builder.advanceLexer(); // ;
            }
            parseExpression(builder); // increment
            if (builder.getTokenType() == BuddhistTypes.RPAREN) {
                builder.advanceLexer(); // )
            }
        }
        if (builder.getTokenType() == BuddhistTypes.LBRACE) {
            parseBlock(builder);
        }
        marker.done(BuddhistTypes.FOR);
    }

    private void parseFunctionStatement(PsiBuilder builder) {
        PsiBuilder.Marker marker = builder.mark();
        builder.advanceLexer(); // FUNCTION
        if (builder.getTokenType() == BuddhistTypes.IDENTIFIER) {
            builder.advanceLexer(); // identifier
        }
        if (builder.getTokenType() == BuddhistTypes.LPAREN) {
            builder.advanceLexer(); // (
            parseParameters(builder);
            if (builder.getTokenType() == BuddhistTypes.RPAREN) {
                builder.advanceLexer(); // )
            }
        }
        if (builder.getTokenType() == BuddhistTypes.LBRACE) {
            parseBlock(builder);
        }
        marker.done(BuddhistTypes.FUNCTION);
    }

    private void parseParameters(PsiBuilder builder) {
        while (builder.getTokenType() == BuddhistTypes.IDENTIFIER) {
            builder.advanceLexer();
            if (builder.getTokenType() == BuddhistTypes.COMMA) {
                builder.advanceLexer();
            } else {
                break;
            }
        }
    }

    private void parseBlock(PsiBuilder builder) {
        PsiBuilder.Marker marker = builder.mark();
        builder.advanceLexer(); // {
        while (builder.getTokenType() != BuddhistTypes.RBRACE && !builder.eof()) {
            parseStatement(builder);
        }
        if (builder.getTokenType() == BuddhistTypes.RBRACE) {
            builder.advanceLexer(); // }
        }
        marker.done(BuddhistTypes.LBRACE);
    }

    private void parseExpression(PsiBuilder builder) {
        parsePrimaryExpression(builder);
        while (isBinaryOperator(builder.getTokenType())) {
            PsiBuilder.Marker marker = builder.mark();
            builder.advanceLexer();
            parsePrimaryExpression(builder);
            marker.done(BuddhistTypes.PLUS); // Use a generic expression type
        }
    }

    private void parsePrimaryExpression(PsiBuilder builder) {
        IElementType tokenType = builder.getTokenType();
        if (tokenType == BuddhistTypes.IDENTIFIER ||
            tokenType == BuddhistTypes.INT ||
            tokenType == BuddhistTypes.FLOAT ||
            tokenType == BuddhistTypes.STRING ||
            tokenType == BuddhistTypes.TRUE ||
            tokenType == BuddhistTypes.FALSE ||
            tokenType == BuddhistTypes.NULL) {
            builder.advanceLexer();
        } else if (tokenType == BuddhistTypes.LPAREN) {
            builder.advanceLexer(); // (
            parseExpression(builder);
            if (builder.getTokenType() == BuddhistTypes.RPAREN) {
                builder.advanceLexer(); // )
            }
        } else if (tokenType == BuddhistTypes.LBRACKET) {
            parseArray(builder);
        } else if (tokenType == BuddhistTypes.LBRACE) {
            parseObject(builder);
        }
    }

    private void parseArray(PsiBuilder builder) {
        PsiBuilder.Marker marker = builder.mark();
        builder.advanceLexer(); // [
        while (builder.getTokenType() != BuddhistTypes.RBRACKET && !builder.eof()) {
            parseExpression(builder);
            if (builder.getTokenType() == BuddhistTypes.COMMA) {
                builder.advanceLexer();
            } else {
                break;
            }
        }
        if (builder.getTokenType() == BuddhistTypes.RBRACKET) {
            builder.advanceLexer(); // ]
        }
        marker.done(BuddhistTypes.LBRACKET);
    }

    private void parseObject(PsiBuilder builder) {
        PsiBuilder.Marker marker = builder.mark();
        builder.advanceLexer(); // {
        while (builder.getTokenType() != BuddhistTypes.RBRACE && !builder.eof()) {
            if (builder.getTokenType() == BuddhistTypes.STRING) {
                builder.advanceLexer(); // key
            }
            if (builder.getTokenType() == BuddhistTypes.COLON || builder.getTokenType() == BuddhistTypes.ARROW) {
                builder.advanceLexer(); // : or =>
            }
            parseExpression(builder); // value
            if (builder.getTokenType() == BuddhistTypes.COMMA) {
                builder.advanceLexer();
            } else {
                break;
            }
        }
        if (builder.getTokenType() == BuddhistTypes.RBRACE) {
            builder.advanceLexer(); // }
        }
        marker.done(BuddhistTypes.LBRACE);
    }

    private boolean isBinaryOperator(IElementType tokenType) {
        return tokenType == BuddhistTypes.PLUS ||
               tokenType == BuddhistTypes.MINUS ||
               tokenType == BuddhistTypes.ASTERISK ||
               tokenType == BuddhistTypes.SLASH ||
               tokenType == BuddhistTypes.MODULO ||
               tokenType == BuddhistTypes.EQ ||
               tokenType == BuddhistTypes.NOT_EQ ||
               tokenType == BuddhistTypes.LT ||
               tokenType == BuddhistTypes.GT ||
               tokenType == BuddhistTypes.LT_EQ ||
               tokenType == BuddhistTypes.GT_EQ ||
               tokenType == BuddhistTypes.AND ||
               tokenType == BuddhistTypes.OR;
    }
}
