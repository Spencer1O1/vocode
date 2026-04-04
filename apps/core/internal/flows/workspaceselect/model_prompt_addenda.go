package workspaceselectflow

// reactNativeExpoRules is appended to scoped-edit and file-create model system prompts.
// Without it, models often emit web React DOM patterns (<button onClick>) in RN/Expo files.
const reactNativeExpoRules = `

React Native / Expo: When activeFile, fullFile, or targetText indicates React Native or Expo — imports from "react-native", "expo", or "expo-router"; components such as View, Text, Image, Pressable, TouchableOpacity, ScrollView, ThemedView, ThemedText, StyleSheet.create; or paths like app/(tabs)/*.tsx — you MUST follow React Native rules, not browser React DOM:
- Never emit HTML intrinsic elements: no <button>, <div>, <span>, <input>, <p>, <a>, etc. (Wrong: <button onClick={...}> — Right: <Pressable onPress={...}> or <TouchableOpacity onPress={...}> or import { Button } from "react-native".)
- For tappable UI use Pressable, TouchableOpacity, or Button from "react-native" (or the same primitives already used in the file). Handlers must be onPress (and other RN touch props), never onClick.
- Keep layout and styling idiomatic for React Native (flex, StyleSheet, existing themed components) unless that section of the file clearly targets web-only JSX.
- Prefer destructured imports from "react" or "react-native"; avoid unnecessary new imports if fullFile already imports the symbol.
- When the user or instruction needs a new symbol, put the new import lines in importLines (not inside replacementText unless the selection is the import area).
`
