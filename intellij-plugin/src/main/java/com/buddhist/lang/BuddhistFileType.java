package com.buddhist.lang;

import com.intellij.openapi.fileTypes.LanguageFileType;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import javax.swing.*;

public class BuddhistFileType extends LanguageFileType {
    public static final BuddhistFileType INSTANCE = new BuddhistFileType();

    private BuddhistFileType() {
        super(BuddhistLanguage.INSTANCE);
    }

    @NotNull
    @Override
    public String getName() {
        return "Buddhist Language";
    }

    @NotNull
    @Override
    public String getDescription() {
        return "Buddhist Language file";
    }

    @NotNull
    @Override
    public String getDefaultExtension() {
        return "bl";
    }

    @Nullable
    @Override
    public Icon getIcon() {
        return null; // You can add a custom icon here
    }
}
