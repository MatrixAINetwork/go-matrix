using System;
using System.Drawing;
using System.Linq;
using System.Windows.Forms;
using Newtonsoft.Json.Linq;
using JsonTreeView.Extensions;

namespace JsonTreeView.Views
{
    public partial class JTokenTreeView : TreeView
    {
        #region >> Fields

        JTokenTreeNode lastDragDropTarget;
        DateTime lastDragOverDateTime;
        Color lastDragDropTargetBackColor;
        readonly TimeSpan dragDropExpandDelay = new TimeSpan(5000000);

        #endregion

        #region >> Constructors

        /// <summary>
        /// Default constructor
        /// </summary>
        public JTokenTreeView()
        {
            InitializeComponent();

            ItemDrag += ItemDragHandler;
            DragEnter += DragEnterHandler;
            DragDrop += DragDropHandler;
            DragOver += DragOverHandler;
        }

        #endregion

        #region >> TreeView

        /// <inheritdoc />
        /// <remarks>
        /// Style change disabling automatic creation of tooltip on each node of the TreeView (no other C# way of doing this).
        /// </remarks>
        protected override CreateParams CreateParams
        {
            get
            {
                var cp = base.CreateParams;
                cp.Style |= 0x80;    // Turn on TVS_NOTOOLTIPS
                return cp;
            }
        }

        #endregion

        /// <summary>
        /// Occurs when the user begins dragging a node.
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        private void ItemDragHandler(object sender, ItemDragEventArgs e)
        {
            var sourceNode = e.Item as JTokenTreeNode;

            if (sourceNode == null)
            {
                return;
            }

            DoDragDrop(e.Item, DragDropEffects.Move | DragDropEffects.Copy);
        }

        /// <summary>
        /// Occurs when a drag-and-drop operation is completed.
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        private void DragDropHandler(object sender, DragEventArgs e)
        {
            if (lastDragDropTarget != null)
            {
                lastDragDropTarget.BackColor = lastDragDropTargetBackColor;
                lastDragDropTarget = null;
            }

            var sourceNode = GetDragDropSourceNode(e);

            if (sourceNode == null)
            {
                MessageBox.Show(@"Drag & Drop Canceled: Unknown Source");

                return;
            }

            var targetNode = GetDragDropTargetNode(e);

            switch (e.Effect)
            {
                case DragDropEffects.Copy:
                    DoDragDropCopy((dynamic)sourceNode, (dynamic)targetNode);
                    break;
                case DragDropEffects.Move:
                    DoDragDropMove((dynamic)sourceNode, (dynamic)targetNode);
                    break;
            }
        }

        #region >> DoDragDropCopy

        /// <summary>
        /// Catches all unmanaged copies.
        /// </summary>
        /// <param name="sourceNode"></param>
        /// <param name="targetNode"></param>
        private void DoDragDropCopy(JTokenTreeNode sourceNode, JTokenTreeNode targetNode)
        {
            MessageBox.Show(@"Drag & Drop: Unmanaged Copy");
        }

        /// <summary>
        /// Copies a JProperty into a JObject as first child.
        /// </summary>
        /// <param name="sourceNode"></param>
        /// <param name="targetNode"></param>
        private void DoDragDropCopy(JPropertyTreeNode sourceNode, JObjectTreeNode targetNode)
        {
            sourceNode.ClipboardCopy();
            targetNode.ClipboardPasteInto();
        }

        /// <summary>
        /// Copies a JValue into a JArray as first child.
        /// </summary>
        /// <param name="sourceNode"></param>
        /// <param name="targetNode"></param>
        private void DoDragDropCopy(JValueTreeNode sourceNode, JArrayTreeNode targetNode)
        {
            sourceNode.ClipboardCopy();
            targetNode.ClipboardPasteInto();
        }

        /// <summary>
        /// Copies a JObject into a JArray as first child.
        /// </summary>
        /// <param name="sourceNode"></param>
        /// <param name="targetNode"></param>
        private void DoDragDropCopy(JObjectTreeNode sourceNode, JArrayTreeNode targetNode)
        {
            sourceNode.ClipboardCopy();
            targetNode.ClipboardPasteInto();
        }

        /// <summary>
        /// Copies a JArray into a JArray as first child.
        /// </summary>
        /// <param name="sourceNode"></param>
        /// <param name="targetNode"></param>
        private void DoDragDropCopy(JArrayTreeNode sourceNode, JArrayTreeNode targetNode)
        {
            sourceNode.ClipboardCopy();
            targetNode.ClipboardPasteInto();
        }

        #endregion

        #region >> DoDragDropMove

        private void DoDragDropMove(JTokenTreeNode sourceNode, JTokenTreeNode targetNode)
        {
            // TODO: Move sourceNode to target
            MessageBox.Show(@"Drag & Drop: Unmanaged Move");
        }

        /// <summary>
        /// Copies a JProperty into a JObject as first child.
        /// </summary>
        /// <param name="sourceNode"></param>
        /// <param name="targetNode"></param>
        private void DoDragDropMove(JPropertyTreeNode sourceNode, JObjectTreeNode targetNode)
        {
            sourceNode.ClipboardCut();
            targetNode.ClipboardPasteInto();
        }

        /// <summary>
        /// Copies a JObject into a JArray as first child.
        /// </summary>
        /// <param name="sourceNode"></param>
        /// <param name="targetNode"></param>
        private void DoDragDropMove(JObjectTreeNode sourceNode, JArrayTreeNode targetNode)
        {
            sourceNode.ClipboardCut();
            targetNode.ClipboardPasteInto();
        }

        #endregion

        /// <summary>
        /// Occurs when an object is dragged into the control's bounds.
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        private void DragEnterHandler(object sender, DragEventArgs e)
        {
        }

        /// <summary>
        /// Occurs when an object is dragged over the control's bounds. 
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        private void DragOverHandler(object sender, DragEventArgs e)
        {
            var targetNode = GetDragDropTargetNode(e);

            if (targetNode == null)
            {
                e.Effect = DragDropEffects.None;

                if (lastDragDropTarget != null)
                {
                    lastDragDropTarget.BackColor = lastDragDropTargetBackColor;
                }

                lastDragDropTarget = null;

                return;
            }

            var keyState = (KeyStates)e.KeyState;
            if (keyState.HasFlag(KeyStates.Control | KeyStates.Shift))
            {
                e.Effect = DragDropEffects.None;
            }
            else if (keyState.HasFlag(KeyStates.Control))
            {
                e.Effect = DragDropEffects.Copy;
            }
            else if (keyState.HasFlag(KeyStates.Shift))
            {
                e.Effect = DragDropEffects.Move;
            }
            else
            {
                e.Effect = DragDropEffects.Move;
            }

            var sourceNode = GetDragDropSourceNode(e);

            if (targetNode == lastDragDropTarget)
            {
                if (!targetNode.IsExpanded && DateTime.Now - lastDragOverDateTime >= dragDropExpandDelay)
                {
                    targetNode.Expand();
                }

                if (IsDragDropValid(sourceNode, targetNode, e.Effect))
                {
                    lastDragDropTargetBackColor = targetNode.BackColor;
                    targetNode.BackColor = Color.BlueViolet;
                }
                else
                {
                    targetNode.BackColor = lastDragDropTargetBackColor;
                    e.Effect = DragDropEffects.None;
                }
            }
            else
            {
                lastDragDropTarget = targetNode;
                lastDragOverDateTime = DateTime.Now;

                if (IsDragDropValid(sourceNode, targetNode, e.Effect))
                {
                    lastDragDropTargetBackColor = targetNode.BackColor;
                    targetNode.BackColor = Color.BlueViolet;
                }
                else
                {
                    targetNode.BackColor = lastDragDropTargetBackColor;
                    e.Effect = DragDropEffects.None;
                }
            }

            if (lastDragDropTarget != null)
            {
                lastDragDropTarget.BackColor = lastDragDropTargetBackColor;
            }
        }

        private static JTokenTreeNode GetDragDropSourceNode(DragEventArgs e)
        {
            return e.Data.GetData(e.Data.GetFormats().FirstOrDefault(), true) as JTokenTreeNode;
        }

        private JTokenTreeNode GetDragDropTargetNode(DragEventArgs e)
        {
            var targetPoint = PointToClient(new Point(e.X, e.Y));
            var targetNode = GetNodeAt(targetPoint) as JTokenTreeNode;

            return targetNode;
        }

        private bool IsDragDropValid(JTokenTreeNode sourceNode, JTokenTreeNode targetNode, DragDropEffects effect)
        {
            if (sourceNode == null || targetNode == null)
            {
                return false;
            }

            if (sourceNode.JTokenTag is JProperty)
            {
                return targetNode.JTokenTag is JObject;
            }

            if (sourceNode.JTokenTag is JObject)
            {
                switch (effect)
                {
                    case DragDropEffects.Copy:
                        return targetNode.JTokenTag is JArray;
                    case DragDropEffects.Move:
                        return !(targetNode.JTokenTag.Parent is JProperty)
                               && targetNode.JTokenTag is JArray;
                }
            }

            if (sourceNode.JTokenTag is JArray)
            {
                switch (effect)
                {
                    case DragDropEffects.Copy:
                        return targetNode.JTokenTag is JArray;
                    case DragDropEffects.Move:
                        return !(targetNode.JTokenTag.Parent is JProperty)
                               && targetNode.JTokenTag is JArray;
                }
            }

            if (sourceNode.JTokenTag is JValue)
            {
                switch (effect)
                {
                    case DragDropEffects.Copy:
                        return targetNode.JTokenTag is JArray;
                    case DragDropEffects.Move:
                        return !(targetNode.JTokenTag.Parent is JProperty)
                               && targetNode.JTokenTag is JArray;
                }
            }

            return false;
        }
    }
}
