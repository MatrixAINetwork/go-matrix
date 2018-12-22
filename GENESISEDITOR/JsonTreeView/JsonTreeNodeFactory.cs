using System.Linq;
using System.Windows.Forms;
using Newtonsoft.Json.Linq;

namespace JsonTreeView
{
    /// <summary>
    /// Factory allowing to create <see cref="TreeNode"/> encapsulating a given <see cref="JToken"/> instance.
    /// </summary>
    sealed class JsonTreeNodeFactory
    {
        /// <summary>
        /// This depth value allow recursive creation of <see cref="TreeNode"/> for each subnode, without any depth limit.
        /// </summary>
        public const int UnlimitedDepth = 0;

        #region >> Create

        /// <summary>
        /// Create a TreeNode and its subtrees for the <paramref name="obj"/> instance by dynamically dispatching to specialized overloads.
        /// </summary>
        /// <param name="obj"></param>
        /// <param name="depth">Depth of tree of TreeNode to create.
        ///     <list>
        ///         <item>UnlimitedDepth: no limit.</item>
        ///         <item>1: wrap only current object.</item>
        ///         <item>other: desired depth.</item>
        ///     </list>
        /// </param>
        /// <returns></returns>
        public static TreeNode Create(dynamic obj, int depth)
        {
            return Create(obj, depth);
        }

        /// <summary>
        /// Create a TreeNode and its subtrees for the <paramref name="obj"/> instance by dynamically dispatching to specialized overloads.
        /// Depth of the tree is unlimited.
        /// </summary>
        /// <param name="obj"></param>
        /// <returns></returns>
        public static TreeNode Create(dynamic obj)
        {
            return Create(obj, UnlimitedDepth);
        }

        /// <summary>
        /// Create a TreeNode and its subtrees for the <paramref name="obj"/> instance beeing a <see cref="JArray"/> instance.
        /// </summary>
        /// <param name="obj"></param>
        /// <param name="depth"></param>
        /// <returns></returns>
        public static TreeNode Create(JArray obj, int depth)
        {
            var node = new JArrayTreeNode(obj);

            if (depth != 1)
            {
                int remainingDepth = (depth == UnlimitedDepth ? UnlimitedDepth : depth - 1);
                node.Nodes.AddRange(obj
                    .Select(o => Create((dynamic)o, remainingDepth))
                    .Cast<TreeNode>()
                    .ToArray()
                    );
            }

            return node;
        }

        /// <summary>
        /// Create a TreeNode and its subtrees for the <paramref name="obj"/> instance beeing a <see cref="JObject"/> instance.
        /// </summary>
        /// <param name="obj"></param>
        /// <param name="depth"></param>
        /// <returns></returns>
        public static TreeNode Create(JObject obj, int depth)
        {
            var node = new JObjectTreeNode(obj);

            if (depth != 1)
            {
                int remainingDepth = (depth == UnlimitedDepth ? UnlimitedDepth : depth - 1);
                node.Nodes.AddRange(obj.Properties()
                    .Select(o => Create(o, remainingDepth))
                    .ToArray()
                    );
            }

            return node;
        }

        /// <summary>
        /// Create a TreeNode and its subtrees for the <paramref name="obj"/> instance beeing a <see cref="JProperty"/> instance.
        /// </summary>
        /// <param name="obj"></param>
        /// <param name="depth"></param>
        /// <returns></returns>
        public static TreeNode Create(JProperty obj, int depth)
        {
            var node = new JPropertyTreeNode(obj);

            if (depth != 1)
            {
                int remainingDepth = (depth == UnlimitedDepth ? UnlimitedDepth : depth - 1);
                node.Nodes.AddRange(obj
                   .Select(o => Create((dynamic)o, remainingDepth))
                   .Cast<TreeNode>()
                   .ToArray()
                   );
            }

            return node;
        }

        /// <summary>
        /// Throw a <see cref="UnattendedJTokenTypeException"/> for the <paramref name="obj"/> instance beeing a <see cref="JToken"/> instance.
        /// This method exists only for safety in case of a new concrete <see cref="JToken"/> instance is implemented in the future.
        /// </summary>
        /// <param name="obj"></param>
        /// <param name="depth"></param>
        /// <returns></returns>
        /// <exception cref="UnattendedJTokenTypeException">Always thrown.</exception>
        private static TreeNode Create(JToken obj, int depth)
        {
            throw new UnattendedJTokenTypeException(obj);
        }

        /// <summary>
        /// Create a TreeNode and its subtrees for the <paramref name="obj"/> instance beeing a <see cref="JValue"/> instance.
        /// </summary>
        /// <param name="obj"></param>
        /// <param name="depth"></param>
        /// <returns></returns>
        public static TreeNode Create(JValue obj, int depth)
        {
            var node = new JValueTreeNode(obj);

            return node;
        }

        #endregion

        #region >> Wrap

        /// <summary>
        /// Create a TreeNode for the <paramref name="obj"/> instance by dynamically dispatching to specialized overloads.
        /// No subtree is created.
        /// </summary>
        /// <param name="obj"></param>
        /// <returns></returns>
        public static TreeNode Wrap(dynamic obj)
        {
            return Create(obj, 1);
        }

        #endregion
    }
}
