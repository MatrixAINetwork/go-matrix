using System;
using System.Windows.Forms;
using Newtonsoft.Json.Linq;
using JsonTreeView.Generic;

namespace JsonTreeView.Extensions
{
    public static class JTokenTreeNodeActions
    {
        /// <summary>
        /// Implementation of "copy" action
        /// </summary>
        /// <param name="node"></param>
        public static JTokenTreeNode ClipboardCopy(this JTokenTreeNode node)
        {
            EditorClipboard<JTokenTreeNode>.Set(node);

            return node;
        }

        /// <summary>
        /// Implementation of "cut" action
        /// </summary>
        /// <param name="node"></param>
        public static JTokenTreeNode ClipboardCut(this JTokenTreeNode node)
        {
            EditorClipboard<JTokenTreeNode>.Set(node, false);

            return node;
        }

        /// <summary>
        /// Implementation of "paste after" action.
        /// </summary>
        /// <param name="node">Reference node for the paste.</param>
        public static void ClipboardPasteAfter(this JTokenTreeNode node)
        {
            node.ClipboardPaste(
                jt => node.JTokenTag.AddAfterSelf(jt),
                n => node.InsertInParent(n, false)
                );
        }

        /// <summary>
        /// Implementation of "paste before" action.
        /// </summary>
        /// <param name="node"></param>
        public static void ClipboardPasteBefore(this JTokenTreeNode node)
        {
            node.ClipboardPaste(
                jt => node.JTokenTag.AddBeforeSelf(jt),
                n => node.InsertInParent(n, true)
                );
        }

        /// <summary>
        /// Implementation of "paste into" action.
        /// </summary>
        /// <param name="node"></param>
        public static void ClipboardPasteInto(this JTokenTreeNode node)
        {
            node.ClipboardPaste(
                jt => ((JContainer)node.JTokenTag).AddFirst(jt),
                n => node.InsertInCurrent(n)
                );
        }

        /// <summary>
        /// Implementation of "paste and replace" action.
        /// </summary>
        /// <param name="node"></param>
        public static void ClipboardPasteReplace(this JTokenTreeNode node)
        {
            node.ClipboardPaste(
                jt => node.JTokenTag.Replace(jt),
                n => node.InsertInParent(n, true)
                );
        }

        /// <summary>
        /// Implementation of "paste" action using 2 delegates for the concrete action on JToken tree and TreeView.
        /// </summary>
        /// <param name="node"></param>
        /// <param name="pasteJTokenImplementation">Implementation of paste action in the JToken tree.</param>
        /// <param name="pasteTreeNodeImplementation">Implementation of paste action in the treeView.</param>
        private static void ClipboardPaste(this JTokenTreeNode node, Action<JToken> pasteJTokenImplementation, Action<TreeNode> pasteTreeNodeImplementation)
        {
            var sourceJTokenTreeNode = EditorClipboard<JTokenTreeNode>.Get();

            var jTokenSource = sourceJTokenTreeNode.JTokenTag.DeepClone();

            try
            {
                pasteJTokenImplementation(jTokenSource);
            }
            catch (Exception exception)
            {
                // If cut was asked, the clipboard is now empty and source should be inserted again in clipboard
                if (EditorClipboard<JTokenTreeNode>.IsEmpty())
                {
                    EditorClipboard<JTokenTreeNode>.Set(sourceJTokenTreeNode, false);
                }

                throw new JTokenTreeNodePasteException(exception);
            }

            var treeView = node.TreeView;
            treeView.BeginUpdate();

            pasteTreeNodeImplementation(JsonTreeNodeFactory.Create(jTokenSource));

            treeView.EndUpdate();

            // If cut was asked, the clipboard is now empty and source should be removed from treeview
            if (EditorClipboard<JTokenTreeNode>.IsEmpty())
            {
                sourceJTokenTreeNode.EditDelete();
            }
        }

        /// <summary>
        /// Implementation of "delete" action.
        /// </summary>
        /// <param name="node"></param>
        public static void EditDelete(this JTokenTreeNode node)
        {
            if (node == null)
            {
                return;
            }

            try
            {
                node.JTokenTag.Remove();
            }
            catch (Exception exception)
            {
                throw new JTokenTreeNodeDeleteException(exception);
            }

            var treeView = node.TreeView;
            treeView.BeginUpdate();

            node.CleanParentTreeNode();

            treeView.EndUpdate();
        }
    }
}
