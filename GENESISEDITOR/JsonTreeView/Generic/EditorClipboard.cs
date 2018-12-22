
namespace JsonTreeView.Generic
{
    /// <summary>
    /// Generic clipboard for <typeparamref name="T"/> based instance.
    /// </summary>
    /// <typeparam name="T"></typeparam>
    class EditorClipboard<T> where T : class
    {
        static readonly EditorClipboard<T> Clipboard = new EditorClipboard<T>();

        #region >> Fields

        T storedData;
        bool persistentState = true;

        #endregion

        /// <summary>
        /// Clear all data stored in the clipboard.
        /// </summary>
        public static void Clear()
        {
            Clipboard.storedData = null;
            Clipboard.persistentState = true;
        }

        /// <summary>
        /// Get the data stored in the clipboard.
        /// If data is set as not persistant then clipboard is automatically cleared.
        /// </summary>
        /// <returns></returns>
        public static T Get()
        {
            if (Clipboard.persistentState)
            {
                return Clipboard.storedData;
            }

            T source = Clipboard.storedData;
            Clear();
            return source;
        }

        /// <summary>
        /// Indicates if a data is stored in the clipboard.
        /// </summary>
        /// <returns></returns>
        public static bool IsEmpty()
        {
            return Clipboard.storedData == null;
        }

        /// <summary>
        /// Insert a persistent data in the clipboard.
        /// </summary>
        /// <param name="data"></param>
        public static void Set(T data)
        {
            Set(data, true);
        }

        /// <summary>
        /// Insert a data in the clipboard by specifying the persistent state.
        /// </summary>
        /// <param name="data"></param>
        /// <param name="persistentState">
        ///     <list>
        ///         <item><c>true</c> means the data will not be removed from clipboard afert a <see cref="Get()"/>.</item>
        ///         <item><c>false</c> means the data will be removed from clipboard afert a <see cref="Get()"/>.</item>
        ///     </list>
        /// </param>
        public static void Set(T data, bool persistentState)
        {
            Clipboard.storedData = data;
            Clipboard.persistentState = persistentState;
        }
    }
}
