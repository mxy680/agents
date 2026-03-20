import { connect } from "framer-api"
import { createInterface } from "readline"

const projectUrl = process.env.FRAMER_PROJECT_URL
const apiKey = process.env.FRAMER_API_KEY

if (!projectUrl || !apiKey) {
  process.stderr.write(JSON.stringify({ error: "FRAMER_PROJECT_URL and FRAMER_API_KEY are required" }) + "\n")
  process.exit(1)
}

let framer = null

async function handleCommand(cmd) {
  try {
    if (!framer && cmd.method !== "disconnect") {
      framer = await connect(projectUrl, apiKey)
    }

    switch (cmd.method) {
      // Project
      case "getProjectInfo":
        return { result: await framer.getProjectInfo() }
      case "getCurrentUser":
        return { result: await framer.getCurrentUser() }
      case "getChangedPaths":
        return { result: await framer.getChangedPaths() }
      case "getChangeContributors":
        return { result: await framer.getChangeContributors(cmd.params?.fromVersion, cmd.params?.toVersion) }

      // Publishing
      case "publish":
        return { result: await framer.publish() }
      case "deploy": {
        const res = await framer.deploy(cmd.params.deploymentId, cmd.params.domains)
        return { result: res }
      }
      case "getDeployments":
        return { result: await framer.getDeployments() }
      case "getPublishInfo":
        return { result: await framer.getPublishInfo() }

      // Collections
      case "getCollections":
        return { result: await framer.getCollections() }
      case "getCollection":
        return { result: await framer.getCollection(cmd.params.id) }
      case "createCollection":
        return { result: await framer.createCollection(cmd.params.name) }
      case "getCollectionFields": {
        const col = await framer.getCollection(cmd.params.id)
        if (!col) return { error: "collection not found" }
        return { result: await col.getFields() }
      }
      case "addCollectionFields": {
        const col = await framer.getCollection(cmd.params.id)
        if (!col) return { error: "collection not found" }
        return { result: await col.addFields(cmd.params.fields) }
      }
      case "removeCollectionFields": {
        const col = await framer.getCollection(cmd.params.id)
        if (!col) return { error: "collection not found" }
        await col.removeFields(cmd.params.fieldIds)
        return { result: { success: true } }
      }
      case "setCollectionFieldOrder": {
        const col = await framer.getCollection(cmd.params.id)
        if (!col) return { error: "collection not found" }
        await col.setFieldOrder(cmd.params.fieldIds)
        return { result: { success: true } }
      }
      case "getCollectionItems": {
        const col = await framer.getCollection(cmd.params.id)
        if (!col) return { error: "collection not found" }
        return { result: await col.getItems() }
      }
      case "addCollectionItems": {
        const col = await framer.getCollection(cmd.params.id)
        if (!col) return { error: "collection not found" }
        await col.addItems(cmd.params.items)
        return { result: { success: true } }
      }
      case "removeCollectionItems": {
        const col = await framer.getCollection(cmd.params.id)
        if (!col) return { error: "collection not found" }
        await col.removeItems(cmd.params.itemIds)
        return { result: { success: true } }
      }
      case "setCollectionItemOrder": {
        const col = await framer.getCollection(cmd.params.id)
        if (!col) return { error: "collection not found" }
        await col.setItemOrder(cmd.params.ids)
        return { result: { success: true } }
      }

      // Managed Collections
      case "getManagedCollections":
        return { result: await framer.getManagedCollections() }
      case "createManagedCollection":
        return { result: await framer.createManagedCollection(cmd.params.name) }
      case "getManagedCollectionFields": {
        const cols = await framer.getManagedCollections()
        const mc = cols.find(c => c.id === cmd.params.id)
        if (!mc) return { error: "managed collection not found" }
        return { result: await mc.getFields() }
      }
      case "setManagedCollectionFields": {
        const cols = await framer.getManagedCollections()
        const mc = cols.find(c => c.id === cmd.params.id)
        if (!mc) return { error: "managed collection not found" }
        await mc.setFields(cmd.params.fields)
        return { result: { success: true } }
      }
      case "getManagedCollectionItemIds": {
        const cols = await framer.getManagedCollections()
        const mc = cols.find(c => c.id === cmd.params.id)
        if (!mc) return { error: "managed collection not found" }
        return { result: await mc.getItemIds() }
      }
      case "addManagedCollectionItems": {
        const cols = await framer.getManagedCollections()
        const mc = cols.find(c => c.id === cmd.params.id)
        if (!mc) return { error: "managed collection not found" }
        await mc.addItems(cmd.params.items)
        return { result: { success: true } }
      }
      case "removeManagedCollectionItems": {
        const cols = await framer.getManagedCollections()
        const mc = cols.find(c => c.id === cmd.params.id)
        if (!mc) return { error: "managed collection not found" }
        await mc.removeItems(cmd.params.itemIds)
        return { result: { success: true } }
      }
      case "setManagedCollectionItemOrder": {
        const cols = await framer.getManagedCollections()
        const mc = cols.find(c => c.id === cmd.params.id)
        if (!mc) return { error: "managed collection not found" }
        await mc.setItemOrder(cmd.params.ids)
        return { result: { success: true } }
      }

      // Nodes
      case "getNode":
        return { result: await framer.getNode(cmd.params.nodeId) }
      case "getChildren":
        return { result: await framer.getChildren(cmd.params.nodeId) }
      case "getParent":
        return { result: await framer.getParent(cmd.params.nodeId) }
      case "getNodesWithType":
        return { result: await framer.getNodesWithType(cmd.params.type) }
      case "createFrameNode":
        return { result: await framer.createFrameNode(cmd.params.attributes, cmd.params.parentId) }
      case "createTextNode":
        return { result: await framer.createTextNode(cmd.params.attributes, cmd.params.parentId) }
      case "createComponentNode":
        return { result: await framer.createComponentNode(cmd.params.name) }
      case "createWebPage":
        return { result: await framer.createWebPage(cmd.params.path) }
      case "createDesignPage":
        return { result: await framer.createDesignPage(cmd.params.name) }
      case "cloneNode":
        return { result: await framer.cloneNode(cmd.params.nodeId) }
      case "removeNodes":
        await framer.removeNodes(cmd.params.nodeIds)
        return { result: { success: true } }
      case "setAttributes":
        return { result: await framer.setAttributes(cmd.params.nodeId, cmd.params.attributes) }
      case "setParent":
        await framer.setParent(cmd.params.nodeId, cmd.params.parentId, cmd.params.index)
        return { result: { success: true } }
      case "getRect":
        return { result: await framer.getRect(cmd.params.nodeId) }
      case "getCanvasRoot":
        return { result: await framer.getCanvasRoot() }

      // Agent
      case "getAgentSystemPrompt":
        return { result: await framer.getAgentSystemPrompt() }
      case "getAgentContext":
        return { result: await framer.getAgentContext(cmd.params?.options) }
      case "readProjectForAgent":
        return { result: await framer.readProjectForAgent(cmd.params.queries, cmd.params?.options) }
      case "applyAgentChanges":
        return { result: await framer.applyAgentChanges(cmd.params.dsl, cmd.params?.options) }

      // Styles
      case "getColorStyles":
        return { result: await framer.getColorStyles() }
      case "getColorStyle":
        return { result: await framer.getColorStyle(cmd.params.id) }
      case "createColorStyle":
        return { result: await framer.createColorStyle(cmd.params.attributes) }
      case "getTextStyles":
        return { result: await framer.getTextStyles() }
      case "getTextStyle":
        return { result: await framer.getTextStyle(cmd.params.id) }
      case "createTextStyle":
        return { result: await framer.createTextStyle(cmd.params.attributes) }

      // Fonts
      case "getFonts":
        return { result: await framer.getFonts() }
      case "getFont":
        return { result: await framer.getFont(cmd.params.family, cmd.params.attributes) }

      // Localization
      case "getLocales":
        return { result: await framer.getLocales() }
      case "getDefaultLocale":
        return { result: await framer.getDefaultLocale() }
      case "createLocale":
        return { result: await framer.createLocale(cmd.params.input) }
      case "getLocaleLanguages":
        return { result: await framer.getLocaleLanguages() }
      case "getLocaleRegions":
        return { result: await framer.getLocaleRegions(cmd.params.languageCode) }
      case "getLocalizationGroups":
        return { result: await framer.getLocalizationGroups() }
      case "setLocalizationData":
        return { result: await framer.setLocalizationData(cmd.params.update) }

      // Redirects
      case "getRedirects":
        return { result: await framer.getRedirects() }
      case "addRedirects":
        return { result: await framer.addRedirects(cmd.params.redirects) }
      case "removeRedirects":
        await framer.removeRedirects(cmd.params.redirectIds)
        return { result: { success: true } }
      case "setRedirectOrder":
        await framer.setRedirectOrder(cmd.params.redirectIds)
        return { result: { success: true } }

      // Code Files
      case "getCodeFiles":
        return { result: await framer.getCodeFiles() }
      case "getCodeFile":
        return { result: await framer.getCodeFile(cmd.params.id) }
      case "createCodeFile":
        return { result: await framer.createCodeFile(cmd.params.name, cmd.params.code, cmd.params.options) }
      case "typecheckCode":
        return { result: await framer.typecheckCode(cmd.params.fileName, cmd.params.content, cmd.params.compilerOptions, cmd.params.sessionId) }

      // Custom Code
      case "getCustomCode":
        return { result: await framer.getCustomCode() }
      case "setCustomCode":
        await framer.setCustomCode(cmd.params)
        return { result: { success: true } }

      // Images & Files
      case "uploadImage":
        return { result: await framer.uploadImage(cmd.params.image) }
      case "uploadFile":
        return { result: await framer.uploadFile(cmd.params.file) }

      // SVG
      case "getVectorSets":
        return { result: await framer.getVectorSets() }

      // Screenshots
      case "screenshot":
        return { result: await framer.screenshot(cmd.params.nodeId, cmd.params.options) }
      case "exportSVG":
        return { result: await framer.exportSVG(cmd.params.nodeId) }

      // Plugin Data
      case "getPluginData":
        return { result: await framer.getPluginData(cmd.params.key) }
      case "setPluginData":
        await framer.setPluginData(cmd.params.key, cmd.params.value)
        return { result: { success: true } }
      case "getPluginDataKeys":
        return { result: await framer.getPluginDataKeys() }

      // Disconnect
      case "disconnect":
        if (framer) {
          await framer.disconnect()
          framer = null
        }
        return { result: { success: true } }

      default:
        return { error: `unknown method: ${cmd.method}` }
    }
  } catch (err) {
    return { error: err.message || String(err) }
  }
}

const rl = createInterface({ input: process.stdin })

rl.on("line", async (line) => {
  const cmd = JSON.parse(line)
  const response = await handleCommand(cmd)
  process.stdout.write(JSON.stringify(response) + "\n")
})

rl.on("close", async () => {
  if (framer) {
    await framer.disconnect()
  }
  process.exit(0)
})
