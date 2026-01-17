# Buddhist Language IntelliJ Plugin

IntelliJ IDEA plugin for Buddhist Language (.bl files) providing syntax highlighting, code completion, and basic language support.

## Features

- ✅ Syntax highlighting for keywords, strings, numbers, comments
- ✅ File type recognition for `.bl` files
- ✅ Brace matching
- ✅ Comment support (line and block comments)
- ✅ Basic parser structure

## Building the Plugin

### Prerequisites

- JDK 17 or higher
- IntelliJ IDEA (for development)
- Gradle (included via Gradle Wrapper)

**Note:** The plugin uses a simple Java-based lexer (`BuddhistLexerSimple`) that doesn't require JFlex. If you want to use the JFlex-based lexer instead, you'll need to install JFlex and uncomment the build task in `build.gradle`.

### Build Steps

1. Build the plugin:
```bash
./gradlew buildPlugin
```

3. Run the plugin in a sandbox IntelliJ instance:
```bash
./gradlew runIde
```

## Installing the Plugin

1. Build the plugin using `./gradlew buildPlugin`
2. The plugin JAR will be in `build/distributions/`
3. In IntelliJ IDEA:
   - Go to `File` → `Settings` → `Plugins`
   - Click the gear icon → `Install Plugin from Disk...`
   - Select the plugin JAR file
   - Restart IntelliJ IDEA

## Development

The plugin structure:

- `src/main/java/com/buddhist/lang/` - Main plugin code
  - `BuddhistLanguage.java` - Language definition
  - `BuddhistFileType.java` - File type registration
  - `lexer/` - Lexer implementation
  - `parser/` - Parser implementation
  - `psi/` - PSI (Program Structure Interface) elements
  - `highlighting/` - Syntax highlighting
- `src/main/resources/META-INF/plugin.xml` - Plugin configuration

## Language Features Supported

- Keywords: `fn`, `let`, `const`, `if`, `else`, `return`, `for`, `while`, `break`, `continue`, etc.
- Literals: integers, floats, strings
- Operators: arithmetic, comparison, logical
- Comments: line (`//`) and block (`/* */`)
- Data structures: arrays, objects/hash maps

## Future Enhancements

- Code completion
- Go to definition
- Find usages
- Refactoring support
- Error highlighting
- Code formatting

## License

Same as the main Buddhist Language project.
