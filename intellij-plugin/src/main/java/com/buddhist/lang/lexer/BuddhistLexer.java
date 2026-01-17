package com.buddhist.lang.lexer;

import com.intellij.lexer.Lexer;

public class BuddhistLexer {
    public static Lexer createLexer() {
        // Use simple lexer implementation (no JFlex required)
        return new BuddhistLexerSimple();
    }
}
