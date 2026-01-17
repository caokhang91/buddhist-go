package com.buddhist.lang.highlighting;

import com.intellij.openapi.editor.colors.TextAttributesKey;
import com.intellij.openapi.fileTypes.SyntaxHighlighter;
import com.intellij.openapi.options.colors.AttributesDescriptor;
import com.intellij.openapi.options.colors.ColorDescriptor;
import com.intellij.openapi.options.colors.ColorSettingsPage;
import com.buddhist.lang.BuddhistFileType;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import javax.swing.*;
import java.util.Map;

public class BuddhistColorSettingsPage implements ColorSettingsPage {
    private static final AttributesDescriptor[] DESCRIPTORS = new AttributesDescriptor[]{
            new AttributesDescriptor("Keyword", BuddhistSyntaxHighlighterColors.KEYWORD),
            new AttributesDescriptor("String", BuddhistSyntaxHighlighterColors.STRING),
            new AttributesDescriptor("Number", BuddhistSyntaxHighlighterColors.NUMBER),
            new AttributesDescriptor("Comment", BuddhistSyntaxHighlighterColors.COMMENT),
            new AttributesDescriptor("Identifier", BuddhistSyntaxHighlighterColors.IDENTIFIER),
            new AttributesDescriptor("Operator", BuddhistSyntaxHighlighterColors.OPERATOR),
    };

    @Nullable
    @Override
    public Icon getIcon() {
        return null;
    }

    @NotNull
    @Override
    public SyntaxHighlighter getHighlighter() {
        return new BuddhistSyntaxHighlighterFactory.BuddhistSyntaxHighlighter();
    }

    @NotNull
    @Override
    public String getDemoText() {
        return "// Buddhist Language Example\n" +
               "let name = \"Buddhist\";\n" +
               "let version = 1.0;\n" +
               "\n" +
               "let fibonacci = fn(n) {\n" +
               "    if (n <= 1) {\n" +
               "        return n;\n" +
               "    }\n" +
               "    return fibonacci(n - 1) + fibonacci(n - 2);\n" +
               "};\n" +
               "\n" +
               "let arr = [1, 2, 3, 4, 5];\n" +
               "let person = {\"name\": \"Buddhist\", \"age\": 1};";
    }

    @Nullable
    @Override
    public Map<String, TextAttributesKey> getAdditionalHighlightingTagToDescriptorMap() {
        return null;
    }

    @NotNull
    @Override
    public AttributesDescriptor[] getAttributeDescriptors() {
        return DESCRIPTORS;
    }

    @NotNull
    @Override
    public ColorDescriptor[] getColorDescriptors() {
        return ColorDescriptor.EMPTY_ARRAY;
    }

    @NotNull
    @Override
    public String getDisplayName() {
        return "Buddhist Language";
    }
}
