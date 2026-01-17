package com.buddhist.lang.psi;

import com.buddhist.lang.BuddhistLanguage;
import com.intellij.extapi.psi.PsiFileBase;
import com.intellij.openapi.fileTypes.FileType;
import com.intellij.psi.FileViewProvider;
import com.buddhist.lang.BuddhistFileType;
import org.jetbrains.annotations.NotNull;

public class BuddhistFile extends PsiFileBase {
    public BuddhistFile(@NotNull FileViewProvider viewProvider) {
        super(viewProvider, BuddhistLanguage.INSTANCE);
    }

    @NotNull
    @Override
    public FileType getFileType() {
        return BuddhistFileType.INSTANCE;
    }
}
