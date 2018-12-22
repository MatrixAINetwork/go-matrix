using System.Windows.Forms;
using Newtonsoft.Json.Linq;
using JsonTreeView.Generic;

namespace JsonTreeView
{
    /// <summary>
    /// Specialized <see cref="TreeNode"/> for handling <see cref="JValue"/> representation in a <see cref="TreeView"/>.
    /// </summary>
    sealed class JValueTreeNode : JTokenTreeNode
    {
        #region >> Properties

        public JValue JValueTag => Tag as JValue;

        #endregion

        #region >> Constructors

        /// <summary>
        /// Initializes a new instance of the <see cref="JValueTreeNode"/> class.
        /// </summary>
        /// <param name="jValue"></param>
        public JValueTreeNode(JToken jValue)
            : base(jValue)
        {
            ContextMenuStrip = SingleInstanceProvider<JValueContextMenuStrip>.Value;
        }

        #endregion

        #region >> JTokenTreeNode

        /// <inheritdoc />
        public override void AfterCollapse()
        {
            base.AfterCollapse();

            Text = GetAbstractTextForTag();
        }

        /// <inheritdoc />
        public override void AfterExpand()
        {
            base.AfterExpand();

            Text = GetAbstractTextForTag();
        }

        #endregion
    }
}
