using System.Collections.Generic;
using System.Linq;
using System.Windows.Forms;
using Newtonsoft.Json.Linq;
using JsonTreeView.Generic;
using JsonTreeView.Json;
using JsonTreeView.Linq;

namespace JsonTreeView
{
    /// <summary>
    /// Specialized <see cref="TreeNode"/> for handling <see cref="JProperty"/> representation in a <see cref="TreeView"/>.
    /// </summary>
    sealed class JPropertyTreeNode : JTokenTreeNode
    {
        #region >> Properties

        public JProperty JPropertyTag => Tag as JProperty;

        #endregion

        #region >> Constructors

        /// <summary>
        /// Initializes a new instance of the <see cref="JPropertyTreeNode"/> class.
        /// </summary>
        public JPropertyTreeNode(JProperty jProperty)
            : base(jProperty)
        {
            ContextMenuStrip = SingleInstanceProvider<JPropertyContextMenuStrip>.Value;
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

            Text = JPropertyTag.Name;
        }

        /// <inheritdoc />
        public override TreeNode AfterJsonTextChange(string jsonString)
        {
            if (CheckEmptyJsonString(jsonString))
            {
                return null;
            }

            // To allow parsing, the partial json string is first enclosed as a json object

            var jTokenRoot = new JTokenRoot("{" + jsonString + "}");

            // Extract the contained JProperties as the JObject was only a container
            // As Json.NET internally clones JToken instances having Parent!=null when inserting in a JContainer,
            // explicitly clones the new JProperties to nullify Parent and to know of the instances
            var jParsedProperties = ((JObject)jTokenRoot.JTokenValue).Properties()
                .Select(p => new JProperty(p))
                .ToList();

            // Update the properties of parent JObject by inserting jParsedProperties and removing edited JProperty
            var jObjectParent = (JObject)JPropertyTag.Parent;

            var jProperties = jObjectParent.Properties()
                .SelectMany(p => ReferenceEquals(p, JPropertyTag) ? jParsedProperties : new List<JProperty> { p })
                .Distinct(new JPropertyEqualityComparer())
                .ToList();
            jObjectParent.ReplaceAll(jProperties);

            // Build a new list of TreeNodes for these JProperties
            var jParsedTreeNodes = jParsedProperties
                .Select(p => JsonTreeNodeFactory.Create(p))
                .Cast<JPropertyTreeNode>()
                .ToList();

            return UpdateTreeNodes(jParsedTreeNodes);
        }

        #endregion

        /// <summary>
        /// Insert or replace a set of <paramref name="newNodes"/>s in current parent nodes.
        /// </summary>
        /// <param name="newNodes"></param>
        /// <returns></returns>
        public TreeNode UpdateTreeNodes(IEnumerable<JPropertyTreeNode> newNodes)
        {
            var treeNodeCollection = Parent?.Nodes ?? TreeView.Nodes;

            var nodeIndex = treeNodeCollection.IndexOf(this);

            newNodes.ForEach(n => treeNodeCollection.Insert(nodeIndex++, n));

            CleanParentTreeNode();

            return newNodes.FirstOrDefault();
        }
    }
}
