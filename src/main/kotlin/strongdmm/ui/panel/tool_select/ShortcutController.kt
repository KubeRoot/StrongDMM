package strongdmm.ui.panel.tool_select

import org.lwjgl.glfw.GLFW
import strongdmm.service.shortcut.ShortcutHandler

class ShortcutController(
    private val state: State
) {
    private val shortcutHandler = ShortcutHandler()

    lateinit var viewController: ViewController

    init {
        shortcutHandler.addShortcut(GLFW.GLFW_KEY_1) { viewController.doSelectTool(state.tools[0]) }
        shortcutHandler.addShortcut(GLFW.GLFW_KEY_2) { viewController.doSelectTool(state.tools[1]) }
        shortcutHandler.addShortcut(GLFW.GLFW_KEY_3) { viewController.doSelectTool(state.tools[2]) }
    }
}
