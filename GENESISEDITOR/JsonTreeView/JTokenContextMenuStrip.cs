using System;
using System.Linq;
using System.Windows.Forms;
using JsonTreeView.Extensions;
using JsonTreeView.Generic;
using JsonTreeView.Linq;
using JsonTreeView.Properties;

namespace JsonTreeView
{
    class JTokenContextMenuStrip : ContextMenuStrip
    {
        /// <summary>
        /// Source <see cref="TreeNode"/> at the origin of this <see cref="ContextMenuStrip"/>
        /// </summary>
        protected JTokenTreeNode JTokenNode;

        protected ToolStripItem CollapseAllToolStripItem;
        protected ToolStripItem ExpandAllToolStripItem;

        protected ToolStripMenuItem EditToolStripItem;

        protected ToolStripItem CopyNodeToolStripItem;
        protected ToolStripItem CutNodeToolStripItem;
        protected ToolStripItem DeleteNodeToolStripItem;
        protected ToolStripItem PasteNodeAfterToolStripItem;
        protected ToolStripItem PasteNodeBeforeToolStripItem;
        protected ToolStripItem PasteNodeReplaceToolStripItem;

        #region >> Constructors

        /// <summary>
        /// Initializes a new instance of the <see cref="JTokenContextMenuStrip"/> class.
        /// </summary>
        public JTokenContextMenuStrip()
        {
            CollapseAllToolStripItem = new ToolStripMenuItem(Resources.CollapseAll, null, CollapseAll_Click);
            ExpandAllToolStripItem = new ToolStripMenuItem(Resources.ExpandAll, null, ExpandAll_Click);

            EditToolStripItem = new ToolStripMenuItem(Resources.Edit);

            CopyNodeToolStripItem = new ToolStripMenuItem(Resources.Copy, null, CopyNode_Click);
            CutNodeToolStripItem = new ToolStripMenuItem(Resources.Cut, null, CutNode_Click);
            DeleteNodeToolStripItem = new ToolStripMenuItem(Resources.DeleteNode, null, DeleteNode_Click);
            PasteNodeAfterToolStripItem = new ToolStripMenuItem(Resources.PasteNodeAfter, null, PasteNodeAfter_Click);
            PasteNodeBeforeToolStripItem = new ToolStripMenuItem(Resources.PasteNodeBefore, null, PasteNodeBefore_Click);
            PasteNodeReplaceToolStripItem = new ToolStripMenuItem(Resources.Replace, null, PasteNodeReplace_Click);

            EditToolStripItem.DropDownItems.Add(CopyNodeToolStripItem);
            EditToolStripItem.DropDownItems.Add(CutNodeToolStripItem);
            EditToolStripItem.DropDownItems.Add(PasteNodeBeforeToolStripItem);
            EditToolStripItem.DropDownItems.Add(PasteNodeAfterToolStripItem);
            EditToolStripItem.DropDownItems.Add(new ToolStripSeparator());
            EditToolStripItem.DropDownItems.Add(PasteNodeReplaceToolStripItem);
            EditToolStripItem.DropDownItems.Add(new ToolStripSeparator());
            EditToolStripItem.DropDownItems.Add(DeleteNodeToolStripItem);

            Items.Add(CollapseAllToolStripItem);
            Items.Add(ExpandAllToolStripItem);
            Items.Add(EditToolStripItem);
        }

        #endregion

        #region >> ContextMenuStrip

        /// <inheritdoc />
        protected override void OnVisibleChanged(EventArgs e)
        {
            if (Visible)
            {
                JTokenNode = FindSourceTreeNode<JTokenTreeNode>();

                // Collapse item shown if node is expanded and has children
                CollapseAllToolStripItem.Visible = JTokenNode.IsExpanded
                    && JTokenNode.Nodes.Cast<TreeNode>().Any();

                // Expand item shown if node if not expanded or has a children not expanded
                ExpandAllToolStripItem.Visible = !JTokenNode.IsExpanded
                    || JTokenNode.Nodes.Cast<TreeNode>().Any(t => !t.IsExpanded);

                // Remove item enabled if it is not the root or the value of a property
                DeleteNodeToolStripItem.Enabled = (JTokenNode.Parent != null)
                    && !(JTokenNode.Parent is JPropertyTreeNode);

                // Cut item enabled if delete is
                CutNodeToolStripItem.Enabled = DeleteNodeToolStripItem.Enabled;

                // Paste items enabled only when a copy or cut operation is pending
                PasteNodeAfterToolStripItem.Enabled = !EditorClipboard<JTokenTreeNode>.IsEmpty()
                    && (JTokenNode.Parent != null)
                    && !(JTokenNode.Parent is JPropertyTreeNode);

                PasteNodeBeforeToolStripItem.Enabled = !EditorClipboard<JTokenTreeNode>.IsEmpty()
                    && (JTokenNode.Parent != null)
                    && !(JTokenNode.Parent is JPropertyTreeNode);

                PasteNodeReplaceToolStripItem.Enabled = !EditorClipboard<JTokenTreeNode>.IsEmpty()
                    && (JTokenNode.Parent != null);
            }

            base.OnVisibleChanged(e);
        }

        #endregion

        /// <summary>
        /// Click event handler for <see cref="CollapseAllToolStripItem"/>.
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        void CollapseAll_Click(Object sender, EventArgs e)
        {
            if (JTokenNode != null)
            {
                JTokenNode.TreeView.BeginUpdate();

                var nodes = JTokenNode.EnumerateNodes().Take(1000);
                foreach (var treeNode in nodes)
                {
                    treeNode.Collapse();
                }

                JTokenNode.TreeView.EndUpdate();
            }
        }

        /// <summary>
        /// Click event handler for <see cref="CopyNodeToolStripItem"/>.
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        void CopyNode_Click(Object sender, EventArgs e)
        {
            JTokenNode.ClipboardCopy();
        }

        /// <summary>
        /// Click event handler for <see cref="CutNodeToolStripItem"/>.
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        void CutNode_Click(Object sender, EventArgs e)
        {
            JTokenNode.ClipboardCut();
        }

        /// <summary>
        /// Click event handler for <see cref="DeleteNodeToolStripItem"/>.
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        void DeleteNode_Click(Object sender, EventArgs e)
        {
            try
            {
                JTokenNode.EditDelete();
            }
            catch (JTokenTreeNodeDeleteException exception)
            {
                MessageBox.Show(exception.InnerException?.Message, Resources.DeletionActionFailed);
            }
        }

        /// <summary>
        /// Click event handler for <see cref="ExpandAllToolStripItem"/>.
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        void ExpandAll_Click(Object sender, EventArgs e)
        {
            if (JTokenNode != null)
            {
                JTokenNode.TreeView.BeginUpdate();

                var nodes = JTokenNode.EnumerateNodes().Take(1000);
                foreach (var treeNode in nodes)
                {
                    treeNode.Expand();
                }

                JTokenNode.TreeView.EndUpdate();
            }
        }

        /// <summary>
        /// Click event handler for <see cref="PasteNodeAfterToolStripItem"/>.
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        void PasteNodeAfter_Click(Object sender, EventArgs e)
        {
            try
            {
                JTokenNode.ClipboardPasteAfter();
            }
            catch (JTokenTreeNodePasteException exception)
            {
                MessageBox.Show(exception.InnerException?.Message, Resources.PasteActionFailed);
            }
        }

        /// <summary>
        /// Click event handler for <see cref="PasteNodeBeforeToolStripItem"/>.
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        void PasteNodeBefore_Click(Object sender, EventArgs e)
        {
            try
            {
                JTokenNode.ClipboardPasteBefore();
            }
            catch (JTokenTreeNodePasteException exception)
            {
                MessageBox.Show(exception.InnerException?.Message, Resources.PasteActionFailed);
            }
        }

        /// <summary>
        /// Click event handler for <see cref="PasteNodeReplaceToolStripItem"/>.
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        void PasteNodeReplace_Click(Object sender, EventArgs e)
        {
            try
            {
                JTokenNode.ClipboardPasteReplace();
            }
            catch (JTokenTreeNodePasteException exception)
            {
                MessageBox.Show(exception.InnerException?.Message, Resources.PasteActionFailed);
            }
        }

        /// <summary>
        /// Identify the Source <see cref="TreeNode"/> at the origin of this <see cref="ContextMenuStrip"/>.
        /// </summary>
        /// <typeparam name="T">Subtype of <see cref="TreeNode"/> to return.</typeparam>
        /// <returns></returns>
        public T FindSourceTreeNode<T>() where T : TreeNode
        {
            var treeView = SourceControl as TreeView;

            return treeView?.SelectedNode as T;
        }
    }
}
