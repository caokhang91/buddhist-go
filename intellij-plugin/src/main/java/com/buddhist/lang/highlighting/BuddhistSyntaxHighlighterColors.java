package com.buddhist.lang.highlighting;

import com.intellij.openapi.editor.DefaultLanguageHighlighterColors;
import com.intellij.openapi.editor.colors.TextAttributesKey;

public class BuddhistSyntaxHighlighterColors {
    public static final TextAttributesKey KEYWORD =
            TextAttributesKey.createTextAttributesKey("BUDDHIST_KEYWORD", DefaultLanguageHighlighterColors.KEYWORD);
    
    public static final TextAttributesKey STRING =
            TextAttributesKey.createTextAttributesKey("BUDDHIST_STRING", DefaultLanguageHighlighterColors.STRING);
    
    public static final TextAttributesKey NUMBER =
            TextAttributesKey.createTextAttributesKey("BUDDHIST_NUMBER", DefaultLanguageHighlighterColors.NUMBER);
    
    public static final TextAttributesKey COMMENT =
            TextAttributesKey.createTextAttributesKey("BUDDHIST_COMMENT", DefaultLanguageHighlighterColors.LINE_COMMENT);
    
    public static final TextAttributesKey IDENTIFIER =
            TextAttributesKey.createTextAttributesKey("BUDDHIST_IDENTIFIER", DefaultLanguageHighlighterColors.IDENTIFIER);
    
    public static final TextAttributesKey OPERATOR =
            TextAttributesKey.createTextAttributesKey("BUDDHIST_OPERATOR", DefaultLanguageHighlighterColors.OPERATION_SIGN);
}
