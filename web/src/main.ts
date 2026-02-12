import { createApp, h } from "vue";
import {
  create,
  NButton,
  NCode,
  NCollapse,
  NCollapseItem,
  NDialogProvider,
  NDrawer,
  NDrawerContent,
  NList,
  NListItem,
  NMessageProvider,
  NScrollbar,
  NTree
} from "naive-ui/es";
import App from "./App.vue";
import "xterm/css/xterm.css";
import "./styles.css";

// bootstrapApp 初始化并挂载应用。
function bootstrapApp() {
  const naive = create({
    components: [
      NButton,
      NCode,
      NCollapse,
      NCollapseItem,
      NDialogProvider,
      NDrawer,
      NDrawerContent,
      NList,
      NListItem,
      NMessageProvider,
      NScrollbar,
      NTree
    ]
  });
  // 使用消息与对话框提供者包裹根组件，确保 useMessage/useDialog 有上层提供者。
  const Root = {
    name: "RootApp",
    setup() {
      // setup 返回渲染函数，确保 App 位于 NMessageProvider 与 NDialogProvider 下。
      return () =>
        h(NMessageProvider, null, {
          default: () => h(NDialogProvider, null, { default: () => h(App) })
        });
    }
  };
  createApp(Root).use(naive).mount("#app");
}

bootstrapApp();
