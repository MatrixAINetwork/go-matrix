using System;
using System.Windows.Forms;
using Newtonsoft.Json.Linq;
using JsonTreeView.Properties;

namespace JsonTreeView
{
    class JArrayContextMenuStrip : JTokenContextMenuStrip
    {
        protected ToolStripMenuItem ArrayToolStripItem;
        protected ToolStripMenuItem InsertArrayToolStripItem;
        protected ToolStripMenuItem InsertObjectToolStripItem;
        protected ToolStripMenuItem InsertValueToolStripItem;

        #region >> Constructors

        public JArrayContextMenuStrip()
        {
            ArrayToolStripItem = new ToolStripMenuItem(Resources.JsonArray);
            InsertArrayToolStripItem = new ToolStripMenuItem(Resources.InsertArray, null, InsertArray_Click);
            InsertObjectToolStripItem = new ToolStripMenuItem(Resources.InsertObject, null, InsertObject_Click);
            InsertValueToolStripItem = new ToolStripMenuItem(Resources.InsertValue, null, InsertValue_Click);

            ArrayToolStripItem.DropDownItems.Add(InsertArrayToolStripItem);
            ArrayToolStripItem.DropDownItems.Add(InsertObjectToolStripItem);
            ArrayToolStripItem.DropDownItems.Add(InsertValueToolStripItem);
            Items.Add(ArrayToolStripItem);
        }

        #endregion

        /// <summary>
        /// Click event handler for <see cref="InsertValueToolStripItem"/>.
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        private void InsertArray_Click(Object sender, EventArgs e)
        {
            InsertJToken(JArray.Parse("[]"));
        }

        /// <summary>
        /// Click event handler for <see cref="InsertValueToolStripItem"/>.
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        private void InsertObject_Click(Object sender, EventArgs e)
        {
            InsertJToken(JObject.Parse("{}"));
        }

        /// <summary>
        /// Click event handler for <see cref="InsertValueToolStripItem"/>.
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        private void InsertValue_Click(Object sender, EventArgs e)
        {
            InsertJToken(JToken.Parse("null"));
        }

        /// <summary>
        /// Add a new <see cref="JToken"/> instance in current <see cref="JArrayTreeNode"/>
        /// </summary>
        /// <param name="newJToken"></param>
        private void InsertJToken(JToken newJToken)
        {
            var jArrayTreeNode = JTokenNode as JArrayTreeNode;

            if (jArrayTreeNode == null)
            {
                return;
            }

            jArrayTreeNode.JArrayTag.AddFirst(newJToken);

            TreeNode newTreeNode = JsonTreeNodeFactory.Create(newJToken);
            jArrayTreeNode.Nodes.Insert(0, newTreeNode);

            jArrayTreeNode.TreeView.SelectedNode = newTreeNode;
        }
    }
}
