using System.Collections.Generic;
using System.Linq;
using System.Windows.Forms;

namespace JsonTreeView.Linq
{
    public static class TreeNodeExtension
    {
        /// <summary>
        /// Enumerates node all its sub nodes (depth first).
        /// </summary>
        /// <param name="node"></param>
        /// <returns></returns>
        public static IEnumerable<TreeNode> EnumerateNodes(this TreeNode node)
        {
            var stack = new Stack<TreeNode>();

            stack.Push(node);
            while (stack.Count > 0)
            {
                var current = stack.Pop();

                yield return current;

                foreach (var child in current.Nodes.Cast<TreeNode>().Reverse())
                {
                    stack.Push(child);
                }
            }
        }
    }
}
