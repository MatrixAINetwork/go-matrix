using System;
using System.Windows.Forms;
using Newtonsoft.Json.Linq;
using JsonTreeView.Properties;

namespace JsonTreeView
{
    class JObjectContextMenuStrip : JTokenContextMenuStrip
    {
        protected ToolStripMenuItem ObjectToolStripItem;
        protected ToolStripMenuItem InsertPropertyAsValueToolStripItem;
        protected ToolStripMenuItem InsertPropertyAsArrayToolStripItem;
        protected ToolStripMenuItem InsertPropertyAsObjectToolStripItem;

        #region >> Constructors

        /// <summary>
        /// Initializes a new instance of the <see cref="JObjectContextMenuStrip"/> class.
        /// </summary>
        public JObjectContextMenuStrip()
        {
            ObjectToolStripItem = new ToolStripMenuItem(Resources.JsonObject);
            InsertPropertyAsValueToolStripItem = new ToolStripMenuItem(Resources.InsertPropertyAsValue, null, InsertProperty_Click);
            InsertPropertyAsArrayToolStripItem = new ToolStripMenuItem(Resources.InsertPropertyAsArray, null, InsertPropertyAsArray_Click);
            InsertPropertyAsObjectToolStripItem = new ToolStripMenuItem(Resources.InsertPropertyAsObject, null, InsertPropertyAsObject_Click);

            ObjectToolStripItem.DropDownItems.Add(InsertPropertyAsValueToolStripItem);
            ObjectToolStripItem.DropDownItems.Add(InsertPropertyAsArrayToolStripItem);
            ObjectToolStripItem.DropDownItems.Add(InsertPropertyAsObjectToolStripItem);
            Items.Add(ObjectToolStripItem);
        }

        #endregion

        private void InsertProperty(object propertyValue)
        {
            var jObjectTreeNode = JTokenNode as JObjectTreeNode;

            if (jObjectTreeNode == null)
            {
                return;
            }

            var newJProperty = new JProperty("name" + DateTime.Now.Ticks, propertyValue);
            jObjectTreeNode.JObjectTag.AddFirst(newJProperty);

            var jPropertyTreeNode = (JPropertyTreeNode)JsonTreeNodeFactory.Create(newJProperty);
            jObjectTreeNode.Nodes.Insert(0, jPropertyTreeNode);

            jObjectTreeNode.TreeView.SelectedNode = jPropertyTreeNode;
        }

        /// <summary>
        /// Click event handler for <see cref="InsertPropertyAsValueToolStripItem"/>.
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        private void InsertProperty_Click(Object sender, EventArgs e)
        {
            InsertProperty("v");
        }

        /// <summary>
        /// Click event handler for <see cref="InsertPropertyAsArrayToolStripItem"/>.
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        private void InsertPropertyAsArray_Click(object sender, EventArgs e)
        {
            InsertProperty(new JArray());
        }

        /// <summary>
        /// Click event handler for <see cref="InsertPropertyAsObjectToolStripItem"/>.
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        private void InsertPropertyAsObject_Click(object sender, EventArgs e)
        {
            InsertProperty(new JObject());
        }
    }
}
