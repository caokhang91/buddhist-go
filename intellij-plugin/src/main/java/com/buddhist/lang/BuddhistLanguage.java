package com.buddhist.lang;

import com.intellij.lang.Language;

public class BuddhistLanguage extends Language {
    public static final BuddhistLanguage INSTANCE = new BuddhistLanguage();

    private BuddhistLanguage() {
        super("BuddhistLanguage");
    }
}
